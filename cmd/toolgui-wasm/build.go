package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	wasmweb "github.com/voilelab/toolgui/toolgui-web/wasm"
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

// buildOpts is what build and serve share.
type buildOpts struct {
	out     string
	pkg     string
	ldflags string // passed to go build as -ldflags, e.g. "-s -w"
}

// parseBuildFlags reads the flags build and serve share. extra registers the
// ones only the caller wants.
func parseBuildFlags(name string, args []string, extra func(*flag.FlagSet)) (buildOpts, error) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	out := flags.String("o", "dist", "directory to write the site into")
	ldflags := flags.String("ldflags", "", "arguments to pass on each go tool link invocation")
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

	return buildOpts{out: *out, pkg: pkg, ldflags: *ldflags}, nil
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

	err = compile(out, opts.pkg, opts.ldflags)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = copyWasmExec(out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	log.Printf("built %s", out)
	return nil
}

// writeFrontend unpack the embedded browser frontend into out.
func writeFrontend(out string) error {
	assets := wasmweb.GetAssets()

	return fs.WalkDir(assets, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		target := filepath.Join(out, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		bs, err := fs.ReadFile(assets, name)
		if err != nil {
			return err
		}

		return os.WriteFile(target, bs, 0o644)
	})
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
