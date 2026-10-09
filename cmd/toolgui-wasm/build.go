package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"flag"
	"html"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	wasmweb "github.com/voilelab/toolgui/toolgui-web/wasm"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// wasmExecDirs are where the toolchain keeps wasm_exec.js: lib/wasm since Go
// 1.24, misc/wasm before it.
var wasmExecDirs = []string{
	filepath.Join("lib", "wasm"),
	filepath.Join("misc", "wasm"),
}

func runBuild(args []string) error {
	opts, err := parseBuildFlags("build", args, nil)
	if err != nil {
		return err
	}

	return build(opts)
}

// defaultManifest is written when the app brings none. It matches
// tgexec.DefaultManifest, without pulling the server frontend into this
// command.
var defaultManifest = map[string]any{
	"name":             "ToolGUI App",
	"short_name":       "ToolGUI App",
	"start_url":        ".",
	"display":          "standalone",
	"theme_color":      "#000000",
	"background_color": "#ffffff",
}

// buildOpts is what build and serve share.
type buildOpts struct {
	out      string
	pkg      string
	ldflags  string // passed to go build as -ldflags, e.g. "-s -w"
	manifest string // json file written as manifest.json
	assets   string // directory copied to assets/
	lazy     string // directory copied to assets/, cached on first use
	icon     string // favicon url written into index.html
	head     string // html file inserted into the head of index.html
	offline  bool   // write sw.js, so the site opens with no network
}

// parseBuildFlags reads the flags build and serve share. extra registers the
// ones only the caller wants.
func parseBuildFlags(name string, args []string, extra func(*flag.FlagSet)) (buildOpts, error) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	out := flags.String("o", "dist", "directory to write the site into")
	ldflags := flags.String("ldflags", "", "arguments to pass on each go tool link invocation")
	manifest := flags.String("manifest", "", "web app manifest json to write as manifest.json")
	assets := flags.String("assets", "", "directory to copy to assets/, e.g. manifest icons")
	lazy := flags.String("lazy-assets", "", "directory to copy to assets/, kept offline only once fetched, e.g. large runtimes")
	icon := flags.String("icon", "", "favicon url for index.html, e.g. assets/favicon.svg")
	head := flags.String("head", "", "html file to insert into the head of index.html")
	offline := flags.Bool("offline", false, "write a service worker, so the site opens with no network")
	if extra != nil {
		extra(flags)
	}

	err := flags.Parse(args)
	if err != nil {
		return buildOpts{}, tgutil.Errorf("%w", err)
	}

	if flags.NArg() > 1 {
		return buildOpts{}, tgutil.NewError("one package at a time")
	}

	pkg := "."
	if flags.NArg() == 1 {
		pkg = flags.Arg(0)
	}

	return buildOpts{
		out:      *out,
		pkg:      pkg,
		ldflags:  *ldflags,
		manifest: *manifest,
		assets:   *assets,
		lazy:     *lazy,
		icon:     *icon,
		head:     *head,
		offline:  *offline,
	}, nil
}

// build assemble the site in out. Existing files are overwritten, and
// anything else in out is left alone: it may be someone's web root.
func build(opts buildOpts) error {
	out := opts.out
	err := os.MkdirAll(out, 0o755)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = writeFrontend(out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	if opts.icon != "" {
		err = writeIcon(filepath.Join(out, "index.html"), opts.icon)
		if err != nil {
			return tgutil.Errorf("icon: %w", err)
		}
	}

	err = writeManifest(out, opts.manifest)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Before -head, so the theme-color meta found is the shipped one.
	if opts.manifest != "" {
		err = writeAppleTags(filepath.Join(out, "index.html"), opts.manifest, opts.head)
		if err != nil {
			return tgutil.Errorf("%w", err)
		}
	}

	if opts.head != "" {
		err = writeHead(filepath.Join(out, "index.html"), opts.head)
		if err != nil {
			return tgutil.Errorf("head: %w", err)
		}
	}

	if opts.assets != "" && opts.lazy != "" {
		err = checkAssetConflict(opts.assets, opts.lazy)
		if err != nil {
			return tgutil.Errorf("%w", err)
		}
	}

	if opts.assets != "" {
		err = writeAssets(opts.assets, out)
		if err != nil {
			return tgutil.Errorf("copy assets: %w", err)
		}
	}

	if opts.lazy != "" {
		err = writeAssets(opts.lazy, out)
		if err != nil {
			return tgutil.Errorf("copy lazy assets: %w", err)
		}
	}

	err = compile(out, opts.pkg, opts.ldflags)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = copyWasmExec(out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	if opts.offline {
		err = writeServiceWorker(out, opts.assets, opts.lazy)
		if err != nil {
			return tgutil.Errorf("service worker: %w", err)
		}
	}

	log.Printf("built %s", out)
	return nil
}

// writeFrontend unpack the embedded browser frontend into out.
func writeFrontend(out string) error {
	return writeFS(wasmweb.GetAssets(), out)
}

// writeFS copy fsys into out, overwriting files already there.
func writeFS(fsys fs.FS, out string) error {
	return fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		target := filepath.Join(out, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		bs, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}

		return os.WriteFile(target, bs, 0o644)
	})
}

// linkTag matches a <link> tag attribute by attribute, since the emoji data
// url the frontend ships has a raw > inside its href. iconRel and iconHref
// pick the icon link and its href out of one.
var (
	linkTag  = regexp.MustCompile(`<link(?:\s+[^\s"'>/=]+(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'>]+))?)*\s*/?>`)
	iconRel  = regexp.MustCompile(`\srel\s*=\s*(?:"icon"|'icon'|icon[\s/>])`)
	iconHref = regexp.MustCompile(`\shref\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'>]+)`)
)

// writeIcon point the <link rel="icon"> in index.html at icon, so the tab
// shows it before any wasm loads.
func writeIcon(name, icon string) error {
	bs, err := os.ReadFile(name)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	page := string(bs)
	for _, loc := range linkTag.FindAllStringIndex(page, -1) {
		link := page[loc[0]:loc[1]]
		if !iconRel.MatchString(link) {
			continue
		}

		href := ` href="` + html.EscapeString(icon) + `"`
		if iconHref.MatchString(link) {
			link = iconHref.ReplaceAllLiteralString(link, href)
		} else {
			link = "<link" + href + link[len("<link"):]
		}

		page = page[:loc[0]] + link + page[loc[1]:]
		return os.WriteFile(name, []byte(page), 0o644)
	}

	return tgutil.Errorf("no <link rel=\"icon\"> in %s", name)
}

// writeAssets copy src to out/assets. An out inside src is refused: the copy
// would land in the tree being walked and copy itself forever.
func writeAssets(src, out string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	absOut, err := filepath.Abs(out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	rel, err := filepath.Rel(absSrc, absOut)
	if err == nil && filepath.IsLocal(rel) {
		return tgutil.Errorf("output %s is inside assets %s", out, src)
	}

	return writeFS(os.DirFS(src), filepath.Join(out, "assets"))
}

// checkAssetConflict refuse -assets and -lazy-assets that both write a path
// under assets/, or where one has a file the other uses as a directory.
func checkAssetConflict(assets, lazy string) error {
	names, err := fileNames(os.DirFS(assets), "")
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	lazyNames, err := fileNames(os.DirFS(lazy), "")
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	files := map[string]bool{}
	dirs := map[string]bool{}
	for _, name := range names {
		files[name] = true
		for dir := path.Dir(name); dir != "."; dir = path.Dir(dir) {
			dirs[dir] = true
		}
	}

	for _, name := range lazyNames {
		if files[name] || dirs[name] {
			return tgutil.Errorf("assets/%s is in both -assets and -lazy-assets", name)
		}

		for dir := path.Dir(name); dir != "."; dir = path.Dir(dir) {
			if files[dir] {
				return tgutil.Errorf("assets/%s is in both -assets and -lazy-assets", dir)
			}
		}
	}

	return nil
}

// writeManifest write src as out/manifest.json. Without src, the default is
// written, unless out already has a manifest.json someone put there.
func writeManifest(out, src string) error {
	dst := filepath.Join(out, "manifest.json")

	if src == "" {
		_, err := os.Stat(dst)
		if err == nil {
			return nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return tgutil.Errorf("%w", err)
		}

		bs, err := tgjson.Marshal(defaultManifest)
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		return os.WriteFile(dst, bs, 0o644)
	}

	bs, err := os.ReadFile(src)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Fail the build, not the browser, on a broken manifest.
	members := map[string]any{}
	err = tgjson.Unmarshal(bs, &members)
	if err != nil {
		return tgutil.Errorf("manifest %s: %w", src, err)
	}

	// null unmarshals fine, but a manifest is an object.
	if members == nil {
		return tgutil.Errorf("manifest %s: not a json object", src)
	}

	return os.WriteFile(dst, bs, 0o644)
}

// compile build pkg for the browser. The go command reports its own errors,
// so its output goes straight through.
func compile(out, pkg, ldflags string) error {
	args := []string{"build", "-o", filepath.Join(out, "app.wasm")}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, pkg)

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return tgutil.Errorf("go build: %w", err)
	}

	return nil
}

// copyWasmExec take wasm_exec.js from the toolchain that just compiled the
// binary. It is not vendored: the shim and the binary have to match.
func copyWasmExec(out string) error {
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return tgutil.Errorf("go env GOROOT: %w", err)
	}

	src, err := findWasmExec(strings.TrimSpace(string(goroot)))
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	bs, err := os.ReadFile(src)
	if err != nil {
		return tgutil.Errorf("read %s: %w", src, err)
	}

	dst := filepath.Join(out, "wasm_exec.js")

	err = os.WriteFile(dst, bs, 0o644)
	if err != nil {
		return tgutil.Errorf("write %s: %w", dst, err)
	}

	return nil
}

func findWasmExec(goroot string) (string, error) {
	for _, dir := range wasmExecDirs {
		name := filepath.Join(goroot, dir, "wasm_exec.js")
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}

	return "", tgutil.Errorf("no wasm_exec.js under %s", goroot)
}

//go:embed sw.js
var swTemplate string

// offlineMeta tells the page to register sw.js. A page without it unregisters
// one an earlier -offline build left.
const offlineMeta = `<meta name="toolgui-sw" content="sw.js" />`

// writeServiceWorker write sw.js caching the files this build wrote, and mark
// index.html to register it. assets and lazy are the -assets and -lazy-assets
// directories, if any; lazy ones are cached when fetched, not on install.
func writeServiceWorker(out, assets, lazy string) error {
	files, err := fileNames(wasmweb.GetAssets(), "")
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	files = append(files, "app.wasm", "wasm_exec.js", "manifest.json")

	if assets != "" {
		names, err := fileNames(os.DirFS(assets), "assets/")
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		files = append(files, names...)
	}

	lazyFiles := []string{}
	if lazy != "" {
		lazyFiles, err = fileNames(os.DirFS(lazy), "assets/")
		if err != nil {
			return tgutil.Errorf("%w", err)
		}
	}

	err = markIndex(filepath.Join(out, "index.html"))
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Hashed after index.html is marked, so the version covers what is served.
	// Lazy files count too, so a new runtime drops the old one's copy.
	sum := sha256.New()
	for _, name := range slices.Concat(files, lazyFiles) {
		bs, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		sum.Write([]byte(name))
		sum.Write([]byte{0})
		sum.Write(bs)
	}

	list, err := urlList(files)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	lazyList, err := urlList(lazyFiles)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	js := strings.Replace(swTemplate, "__VERSION__", hex.EncodeToString(sum.Sum(nil))[:16], 1)
	js = strings.Replace(js, "__FILES__", list, 1)
	js = strings.Replace(js, "__LAZY__", lazyList, 1)

	return os.WriteFile(filepath.Join(out, "sw.js"), []byte(js), 0o644)
}

// urlList marshal names as a json array of urls. Escaped, so a # or ? in a
// name stays part of the path.
func urlList(names []string) (string, error) {
	urls := make([]string, 0, len(names))
	for _, name := range names {
		segs := strings.Split(name, "/")
		for i, seg := range segs {
			segs[i] = url.PathEscape(seg)
		}

		urls = append(urls, strings.Join(segs, "/"))
	}

	bs, err := tgjson.Marshal(urls)
	if err != nil {
		return "", tgutil.Errorf("%w", err)
	}

	return string(bs), nil
}

// fileNames list the files in fsys, each behind prefix.
func fileNames(fsys fs.FS, prefix string) ([]string, error) {
	names := []string{}
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			names = append(names, prefix+name)
		}

		return nil
	})
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return names, nil
}

// markIndex add offlineMeta to the head of index.html.
func markIndex(name string) error {
	return insertHead(name, offlineMeta)
}

// writeHead insert the html in src into the head of index.html.
func writeHead(index, src string) error {
	bs, err := os.ReadFile(src)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return insertHead(index, strings.TrimSpace(string(bs)))
}

// insertHead put snippet in the head of the file, unless already there.
func insertHead(name, snippet string) error {
	bs, err := os.ReadFile(name)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	html := string(bs)
	if strings.Contains(html, snippet) {
		return nil
	}

	html, ok := tgutil.InsertHead(html, snippet)
	if !ok {
		return tgutil.Errorf("no <head> in %s", name)
	}

	return os.WriteFile(name, []byte(html), 0o644)
}

// appleManifest is what writeAppleTags reads from -manifest.
type appleManifest struct {
	ShortName  string `json:"short_name"`
	ThemeColor string `json:"theme_color"`
	Icons      []struct {
		Src   string `json:"src"`
		Type  string `json:"type"`
		Sizes string `json:"sizes"`
	} `json:"icons"`
}

// metaTag matches a <meta> tag the way linkTag does a <link>.
var metaTag = regexp.MustCompile(`<meta(?:\s+[^\s"'>/=]+(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'>]+))?)*\s*/?>`)

// writeAppleTags add the head tags iOS reads in place of the manifest:
// apple-touch-icon, apple-mobile-web-app-title and theme-color. A tag the
// -head file already has is left to it.
func writeAppleTags(index, manifest, head string) error {
	bs, err := os.ReadFile(manifest)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	m := appleManifest{}
	err = tgjson.Unmarshal(bs, &m)
	if err != nil {
		return tgutil.Errorf("manifest %s: %w", manifest, err)
	}

	headHTML := ""
	if head != "" {
		bs, err := os.ReadFile(head)
		if err != nil {
			return tgutil.Errorf("%w", err)
		}
		headHTML = string(bs)
	}

	tags := []string{}
	if !hasAttr(headHTML, "rel", "apple-touch-icon") {
		icon := touchIcon(m)
		if icon == "" {
			log.Printf("warning: no png icon in %s, so no apple-touch-icon; iOS takes no svg", manifest)
		} else {
			tags = append(tags, `<link rel="apple-touch-icon" href="`+html.EscapeString(icon)+`" />`)
		}
	}

	if m.ShortName != "" && !hasAttr(headHTML, "name", "apple-mobile-web-app-title") {
		tags = append(tags, `<meta name="apple-mobile-web-app-title" content="`+html.EscapeString(m.ShortName)+`" />`)
	}

	if len(tags) > 0 {
		err = insertHead(index, strings.Join(tags, "\n"))
		if err != nil {
			return tgutil.Errorf("%w", err)
		}
	}

	// -head's theme-color replaces the shipped one, not joins it.
	if hasAttr(headHTML, "name", "theme-color") {
		return writeThemeColor(index, "")
	}

	if m.ThemeColor == "" {
		return nil
	}

	return writeThemeColor(index, m.ThemeColor)
}

// hasAttr report whether page has a tag with attr set to value.
func hasAttr(page, attr, value string) bool {
	re := regexp.MustCompile(`(?i)\s` + attr + `\s*=\s*["']?` + regexp.QuoteMeta(value) + `(?:["'\s/>]|$)`)
	return re.MatchString(page)
}

// touchIcon pick the png icon closest to the 180px iOS wants; on a tie, the
// larger, since scaling down looks better. "" if there is no png.
func touchIcon(m appleManifest) string {
	best, bestDist, bestSize := "", -1, 0
	for _, icon := range m.Icons {
		if !isPNG(icon.Type, icon.Src) {
			continue
		}

		// No usable size ranks last, but still beats no icon.
		dist, size := 1<<30, 0
		for _, s := range strings.Fields(icon.Sizes) {
			w, h, ok := strings.Cut(strings.ToLower(s), "x")
			n, err := strconv.Atoi(w)
			_, herr := strconv.Atoi(h)
			if !ok || err != nil || herr != nil || n <= 0 {
				continue
			}

			d := max(n-180, 180-n)
			if d < dist || d == dist && n > size {
				dist, size = d, n
			}
		}

		if bestDist < 0 || dist < bestDist || dist == bestDist && size > bestSize {
			best, bestDist, bestSize = icon.Src, dist, size
		}
	}

	return best
}

// isPNG go by the type, or by the extension when there is none.
func isPNG(typ, src string) bool {
	if typ != "" {
		return strings.EqualFold(typ, "image/png")
	}

	src, _, _ = strings.Cut(src, "#")
	src, _, _ = strings.Cut(src, "?")
	return strings.EqualFold(path.Ext(src), ".png")
}

// writeThemeColor set the theme-color meta in index.html to color, adding
// one if there is none. An empty color removes it.
func writeThemeColor(index, color string) error {
	bs, err := os.ReadFile(index)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	tag := ""
	if color != "" {
		tag = `<meta name="theme-color" content="` + html.EscapeString(color) + `" />`
	}

	page := string(bs)
	for _, loc := range metaTag.FindAllStringIndex(page, -1) {
		if hasAttr(page[loc[0]:loc[1]], "name", "theme-color") {
			page = page[:loc[0]] + tag + page[loc[1]:]
			return os.WriteFile(index, []byte(page), 0o644)
		}
	}

	if tag == "" {
		return nil
	}

	return insertHead(index, tag)
}
