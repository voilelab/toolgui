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
	out, pkg, err := parseBuildFlags("build", args, nil)
	if err != nil {
		return err
	}

	return build(out, pkg)
}

// parseBuildFlags reads the flags build and serve share. extra registers the
// ones only the caller wants.
func parseBuildFlags(name string, args []string, extra func(*flag.FlagSet)) (string, string, error) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	out := flags.String("o", "dist", "directory to write the site into")
	if extra != nil {
		extra(flags)
	}

	err := flags.Parse(args)
	if err != nil {
		return "", "", tgutil.Errorf("%w", err)
	}

	if flags.NArg() > 1 {
		return "", "", tgutil.NewError("one package at a time")
	}

	pkg := "."
	if flags.NArg() == 1 {
		pkg = flags.Arg(0)
	}

	return *out, pkg, nil
}

// build assemble the site in out. Existing files are overwritten, and
// anything else in out is left alone: it may be someone's web root.
func build(out, pkg string) error {
	err := os.MkdirAll(out, 0o755)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = writeFrontend(out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = compile(out, pkg)
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
	return fs.WalkDir(wasmweb.GetAssets(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		target := filepath.Join(out, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		bs, err := fs.ReadFile(wasmweb.GetAssets(), name)
		if err != nil {
			return err
		}

		return os.WriteFile(target, bs, 0o644)
	})
}

// compile build pkg for the browser. The go command reports its own errors,
// so its output goes straight through.
func compile(out, pkg string) error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(out, "app.wasm"), pkg)
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
		return tgutil.Errorf("%w", err)
	}

	err = os.WriteFile(filepath.Join(out, "wasm_exec.js"), bs, 0o644)
	if err != nil {
		return tgutil.Errorf("%w", err)
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
