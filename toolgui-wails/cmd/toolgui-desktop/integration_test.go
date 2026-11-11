package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Opt in after building real frontend assets. This installs the tool in a
// separate consumer module, then builds and runs a real Wails executable with
// Node and Yarn absent from PATH. The probe exits before opening a window.
func TestExternalProject(t *testing.T) {
	if os.Getenv("TOOLGUI_DESKTOP_INTEGRATION") != "1" {
		t.Skip("set TOOLGUI_DESKTOP_INTEGRATION=1 after task asset_wails")
	}
	source, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	desktop := filepath.Clean(filepath.Join(source, "..", ".."))
	if err := checkAssets(desktop); err != nil {
		t.Fatal(err)
	}
	consumer := t.TempDir()
	writeTestFile(t, filepath.Join(consumer, "go.mod"), "module example.org/desktop-consumer\n\ngo 1.27.4\n\nrequire "+desktopImport+" v0.0.0\n\nreplace "+desktopImport+" => "+strconv.Quote(filepath.ToSlash(desktop))+"\nreplace github.com/voilelab/toolgui => "+strconv.Quote(filepath.ToSlash(filepath.Dir(desktop)))+"\n")
	main, err := os.ReadFile(filepath.Join(source, "testdata", "app", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(consumer, "main.go"), string(main))
	// A release consumer gets these checksums from go get. This unreleased
	// fixture uses local replacements and the repository's verified sums.
	var sums []byte
	for _, moduleDir := range []string{desktop, filepath.Dir(desktop)} {
		data, err := os.ReadFile(filepath.Join(moduleDir, "go.sum"))
		if err != nil {
			t.Fatal(err)
		}
		sums = append(sums, data...)
	}
	writeTestFile(t, filepath.Join(consumer, "go.sum"), string(sums))
	t.Setenv("GOWORK", "off")
	t.Setenv("GOOS", runtime.GOOS)
	t.Setenv("GOARCH", runtime.GOARCH)
	t.Setenv("GOFLAGS", "")
	// The unreleased command is supplied by the local replacements above.
	// Prevent go get from probing remote module prefixes for a newer release.
	t.Setenv("GOPROXY", "off")
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	runExternal(t, consumer, goBin, "get", desktopImport+"@v0.0.0")
	runExternal(t, consumer, goBin, "get", "-tool", desktopImport+"/cmd/toolgui-desktop@v0.0.0")
	goroot := strings.TrimSpace(string(runExternal(t, consumer, goBin, "env", "GOROOT")))
	// Keep only the compiler tools needed by Go/cgo. Dependencies have been
	// resolved above; no frontend package manager is available for any build.
	t.Setenv("PATH", compilerOnlyPath(t))
	t.Setenv("GOROOT", goroot)
	t.Setenv("GOFLAGS", "-tags=externalfixture")
	t.Setenv("TOOLGUI_DESKTOP_PROBE", "1")
	for _, name := range []string{"node", "yarn"} {
		if _, err := exec.LookPath(name); err == nil {
			t.Fatalf("%s is still available on the test PATH", name)
		}
	}
	flags := []string{"-tags", "cli_fixture", "-ldflags", "-X 'main.version=release candidate'"}
	output := filepath.Join(consumer, "output with spaces", executableName("example.org/mytool", runtime.GOOS))
	args := append([]string{"tool", "toolgui-desktop", "build", "-o", output}, flags...)
	runExternal(t, consumer, goBin, args...)
	data := runExternal(t, consumer, output, "argument with spaces", "-flag")
	probe := readProbe(t, data)
	builtTags := strings.Split(probe.Tags, ",")
	if probe.Version != "release candidate" || !slices.Contains(builtTags, "desktop") || !slices.Contains(builtTags, "production") {
		t.Fatalf("missing build flags: %+v", probe)
	}
	if !reflect.DeepEqual(probe.Args, []string{"argument with spaces", "-flag"}) {
		t.Fatalf("arguments changed: %v", probe.Args)
	}
	args = append([]string{"tool", "toolgui-desktop", "run"}, flags...)
	args = append(args, ".", "--", "from run", "-input")
	probe = readProbe(t, runExternal(t, consumer, goBin, args...))
	if !reflect.DeepEqual(probe.Args, []string{"from run", "-input"}) {
		t.Fatalf("run arguments changed: %v", probe.Args)
	}
	if _, err := os.Stat(probe.Exe); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("run did not remove temporary executable %s: %v", probe.Exe, err)
	}
	wantCWD, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatal(err)
	}
	gotCWD, err := filepath.EvalSymlinks(probe.CWD)
	if err != nil || gotCWD != wantCWD {
		t.Fatalf("run changed working directory: %s, want %s (%v)", probe.CWD, wantCWD, err)
	}
	args = append([]string{"tool", "toolgui-desktop", "run"}, flags...)
	args = append(args, ".", "--", "exit")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, goBin, args...)
	cmd.Dir = consumer
	var exitErr *exec.ExitError
	if err := cmd.Run(); !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
		t.Fatalf("app exit code not preserved: %v", err)
	}
	// Also cover the default build/bin/<package> location.
	args = append([]string{"tool", "toolgui-desktop", "build"}, flags...)
	runExternal(t, consumer, goBin, args...)
	defaultOutput := filepath.Join(consumer, "build", "bin", executableName("example.org/desktop-consumer", runtime.GOOS))
	if _, err := os.Stat(defaultOutput); err != nil {
		t.Fatalf("default output missing: %v", err)
	}
	if os.Getenv("TOOLGUI_DESKTOP_GUI_SMOKE") == "1" {
		t.Setenv("TOOLGUI_DESKTOP_PROBE", "0")
		if data := runExternalWithTimeout(t, 30*time.Second, consumer, defaultOutput); !bytes.Contains(data, []byte("desktop frontend started")) {
			t.Fatalf("frontend did not start: %s", data)
		}
	}
}

type appProbe struct {
	Version, Tags, Exe, CWD string
	Args                    []string
}

func readProbe(t *testing.T, data []byte) appProbe {
	t.Helper()
	var probe appProbe
	if err := json.Unmarshal(data, &probe); err != nil {
		t.Fatalf("decode probe %s: %v", data, err)
	}
	return probe
}

func runExternal(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	return runExternalWithTimeout(t, 5*time.Minute, dir, name, args...)
}

func runExternalWithTimeout(t *testing.T, timeout time.Duration, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s\n%s", name, args, err, &stdout, &stderr)
	}
	return stdout.Bytes()
}

func compilerOnlyPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"go", "cc", "gcc", "c++", "g++", "clang", "clang++", "as", "ld", "ar", "pkg-config", "xcrun", "xcodebuild", "dsymutil", "sw_vers", "uname", "git"} {
		original, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if runtime.GOOS == "windows" {
			// Windows builds use the Go linker and need no C toolchain.
			if name == "go" {
				data, err := os.ReadFile(original)
				if err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, filepath.Join(dir, "go.exe"), string(data))
			}
			continue
		}
		wrapper := "#!/bin/sh\nexec '" + strings.ReplaceAll(original, "'", "'\"'\"'") + "' \"$@\"\n"
		name = filepath.Join(dir, name)
		writeTestFile(t, name, wrapper)
		if err := os.Chmod(name, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
