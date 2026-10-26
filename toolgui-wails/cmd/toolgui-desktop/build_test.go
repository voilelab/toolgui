package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestCompilerFlags(t *testing.T) {
	tests := []struct {
		name, target, goflags, tags, ldflags, webkit string
		console                                      bool
		installed                                    []string
		wantTags                                     []string
		wantLD, wantErr                              string
	}{
		{name: "macOS preserves custom tags and linker values", target: "darwin", goflags: `'-tags=fromenv shared' '-ldflags=-X main.version=env'`, tags: "custom,shared", wantTags: []string{"desktop", "production", "fromenv", "shared", "custom"}, wantLD: "-X main.version=env"},
		{name: "explicit linker flags override GOFLAGS", target: "darwin", goflags: `'-ldflags=-X main.version=env'`, ldflags: "-X main.version=cli", wantTags: []string{"desktop", "production"}, wantLD: "-X main.version=cli"},
		{name: "deduplicate required tags", target: "darwin", tags: "desktop,production,custom", wantTags: []string{"desktop", "production", "custom"}},
		{name: "Windows GUI", target: "windows", ldflags: "-X main.version=test", wantTags: []string{"desktop", "production"}, wantLD: "-X main.version=test -H=windowsgui"},
		{name: "Windows console", target: "windows", console: true, wantTags: []string{"desktop", "production"}},
		{name: "Linux prefers 4.1", target: "linux", installed: []string{"4.0", "4.1"}, wantTags: []string{"desktop", "production", "webkit2_41"}},
		{name: "Linux falls back to 4.0", target: "linux", installed: []string{"4.0"}, wantTags: []string{"desktop", "production"}},
		{name: "Linux explicit 4.0", target: "linux", installed: []string{"4.0", "4.1"}, webkit: "4.0", wantTags: []string{"desktop", "production"}},
		{name: "Linux explicit 4.1 unavailable", target: "linux", installed: []string{"4.0"}, webkit: "4.1", wantErr: "cannot find"},
		{name: "Linux tag selects 4.1", target: "linux", installed: []string{"4.0", "4.1"}, tags: "webkit2_41", wantTags: []string{"desktop", "production", "webkit2_41"}},
		{name: "Linux conflicting tag", target: "linux", webkit: "4.0", tags: "webkit2_41", wantErr: "conflicts"},
		{name: "Linux missing dependencies", target: "linux", wantErr: "install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev"},
		{name: "Linux 4.1 inherited from GOFLAGS", target: "linux", goflags: "-tags=webkit2_41", installed: []string{"4.0"}, wantErr: "cannot find"},
		{name: "reject dev tag", target: "darwin", tags: "dev", wantErr: "conflicts"},
		{name: "reject inherited bindings tag", target: "darwin", goflags: "-tags=bindings", wantErr: "conflicts"},
		{name: "reject Linux selection on macOS", target: "darwin", webkit: "4.1", wantErr: "Linux-only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.webkit == "" {
				tt.webkit = "auto"
			}
			exists := func(packages ...string) bool {
				if len(packages) != 2 || packages[0] != "gtk+-3.0" {
					t.Fatalf("pkg-config must check GTK and WebKit together: %v", packages)
				}
				return slices.Contains(tt.installed, strings.TrimPrefix(packages[1], "webkit2gtk-"))
			}
			tags, ld, err := compilerFlags(buildOptions{tags: tt.tags, ldflags: tt.ldflags, webkit: tt.webkit, console: tt.console}, goEnvironment{GOOS: tt.target, GOFLAGS: tt.goflags}, exists)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(tags, tt.wantTags) || ld != tt.wantLD {
				t.Fatalf("got tags %v, ldflags %q, error %v; want %v, %q", tags, ld, err, tt.wantTags, tt.wantLD)
			}
		})
	}
}

func TestCompilerEnvironmentPreservesCGO(t *testing.T) {
	original := []string{"PATH=/tools", "CGO_CFLAGS=-O2", "CGO_LDFLAGS=-L/custom -lcustom", "OTHER=value"}
	got := compilerEnvironment(original, "darwin")
	if !slices.Contains(got, "CGO_LDFLAGS=-L/custom -lcustom -framework UniformTypeIdentifiers") || !slices.Contains(got, "CGO_CFLAGS=-O2") {
		t.Fatalf("lost user flags: %v", got)
	}
	if original[2] != "CGO_LDFLAGS=-L/custom -lcustom" {
		t.Fatal("modified caller's environment")
	}
	if got := compilerEnvironment(original, "linux"); !reflect.DeepEqual(got, original) {
		t.Fatalf("changed Linux environment: %v", got)
	}
	if got := compilerEnvironment(nil, "darwin"); !slices.Contains(got, "CGO_LDFLAGS=-framework UniformTypeIdentifiers") {
		t.Fatalf("missing macOS framework: %v", got)
	}
}

// Wails v2.15.0 selects download.go by excluding the other strategies;
// wv2runtime.download is not required to enable its bootstrapper flow.
func TestWindowsWebView2Strategy(t *testing.T) {
	for _, strategy := range []string{"", "wv2runtime.error"} {
		t.Run(strategy, func(t *testing.T) {
			tags, _, err := compilerFlags(buildOptions{tags: strategy, webkit: "auto"}, goEnvironment{GOOS: "windows"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "list", "-f", "{{join .GoFiles \",\"}}", "-tags", strings.Join(tags, ","), "github.com/wailsapp/wails/v2/internal/wv2installer")
			cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0", "GOFLAGS=")
			// Keep stderr out of the parsed list; it may contain "go: downloading" lines.
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			data, err := cmd.Output()
			if err != nil {
				t.Fatalf("go list: %v\n%s", err, &stderr)
			}
			files := strings.Split(strings.TrimSpace(string(data)), ",")
			want := "download.go"
			if strategy != "" {
				want = "error.go"
			}
			for _, file := range []string{"download.go", "browser.go", "embed.go", "error.go"} {
				if slices.Contains(files, file) != (file == want) {
					t.Fatalf("strategy %q selected %v, want only %s", strategy, files, want)
				}
			}
		})
	}
}

func TestCheckTarget(t *testing.T) {
	for _, tt := range []struct {
		env  goEnvironment
		want string
	}{
		{goEnvironment{GOOS: "darwin", GOARCH: "arm64", CGO_ENABLED: "1"}, ""},
		{goEnvironment{GOOS: "windows", GOARCH: "arm64"}, "cross compilation"},
		{goEnvironment{GOOS: "darwin", GOARCH: "amd64", CGO_ENABLED: "1"}, "cross compilation"},
		{goEnvironment{GOOS: "darwin", GOARCH: "arm64", CGO_ENABLED: "0"}, "CGO_ENABLED=1"},
	} {
		err := checkTarget(tt.env, "darwin", "arm64")
		if tt.want == "" && err != nil || tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
			t.Errorf("env %+v: error %v, want %q", tt.env, err, tt.want)
		}
	}
	if err := checkTarget(goEnvironment{GOOS: "windows", GOARCH: "amd64"}, "windows", "amd64"); err != nil {
		t.Fatalf("Windows does not require cgo: %v", err)
	}
}

func TestPackageSelection(t *testing.T) {
	for _, tt := range []struct {
		name     string
		packages []listedPackage
		wantErr  string
	}{
		{"external module", []listedPackage{{ImportPath: desktopImport, Dir: "/module/cache/desktop"}, {ImportPath: "example.org/app", Name: "main", Match: []string{"."}}}, ""},
		{"no desktop dependency", []listedPackage{{ImportPath: "example.org/app", Name: "main", Match: []string{"."}}}, "does not import"},
		{"library", []listedPackage{{Name: "library", Match: []string{"."}}}, "expected main"},
		{"wildcard", []listedPackage{{Name: "main", Match: []string{"./..."}}, {Name: "main", Match: []string{"./..."}}}, "matched 2"},
		{"no matches", nil, "matched 0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var data bytes.Buffer
			for _, pkg := range tt.packages {
				if err := json.NewEncoder(&data).Encode(pkg); err != nil {
					t.Fatal(err)
				}
			}
			_, dir, err := readPackages(&data)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %v, want %q", err, tt.wantErr)
				}
			} else if err != nil || dir != "/module/cache/desktop" {
				t.Fatalf("dir %q, error %v", dir, err)
			}
		})
	}
}

func TestRejectMissingAndPlaceholderAssets(t *testing.T) {
	desktop := t.TempDir()
	if err := checkAssets(desktop); err == nil || !strings.Contains(err.Error(), "tagged toolgui-wails release") {
		t.Fatalf("missing assets: %v", err)
	}
	dir := filepath.Join(desktop, "frontend", "dist")
	writeTestFile(t, filepath.Join(dir, "index.html"), "<title>toolgui: assets not built</title>")
	if err := checkAssets(desktop); err == nil || !strings.Contains(err.Error(), "placeholders") {
		t.Fatalf("placeholder: %v", err)
	}
	writeTestFile(t, filepath.Join(dir, "index.html"), "<script src='/assets/app.js'></script>")
	if err := checkAssets(desktop); err == nil || !strings.Contains(err.Error(), "JavaScript is missing") {
		t.Fatalf("incomplete build: %v", err)
	}
	writeTestFile(t, filepath.Join(dir, "assets", "app.js"), "console.log('built')")
	if err := checkAssets(desktop); err != nil {
		t.Fatal(err)
	}
}

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions("run", []string{"-tags", "custom", "./cmd/app", "--", "-input", "file with spaces"}, io.Discard)
	if err != nil || opts.pkg != "./cmd/app" || opts.tags != "custom" || !reflect.DeepEqual(opts.appArgs, []string{"-input", "file with spaces"}) {
		t.Fatalf("options %+v, error %v", opts, err)
	}
	for _, tt := range []struct {
		mode string
		args []string
	}{
		{"build", []string{"./one", "./two"}},
		{"build", []string{"./app", "-o", "out"}},
		{"build", []string{"--", "arg"}},
		{"build", []string{"main.go"}},
		{"build", []string{"-webkit", "5"}},
		{"run", []string{"-o", "out"}},
	} {
		if _, err := parseOptions(tt.mode, tt.args, io.Discard); err == nil {
			t.Errorf("accepted %s %v", tt.mode, tt.args)
		}
	}
}

func TestHelpDoesNotNeedBuildTools(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, args := range [][]string{{"help"}, {"--help"}, {"build", "-h"}, {"run", "-h"}} {
		var stdout, stderr bytes.Buffer
		if code := command(args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String()+stderr.String(), "toolgui-desktop build") {
			t.Fatalf("%v: code %d, stdout %s, stderr %s", args, code, &stdout, &stderr)
		}
	}
}

func writeTestFile(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
