// Command wasmtest runs a js/wasm test binary in a dedicated Web Worker of a
// headless Chrome, and is meant to be handed to `go test -exec`:
//
//	go build -o /tmp/wasmtest ./scripts/wasmtest
//	GOOS=js GOARCH=wasm go test -exec /tmp/wasmtest ./toolgui/...
//
// A browser is where the tests of a browser build belong, and for the ones
// that reach for the origin private file system it is the only place that will
// do: its synchronous file handles exist in a dedicated worker and nowhere
// else, which is the same reason [tgwasm] runs the app in one.
//
// The worker's stdout is relayed back over HTTP and printed here, and the
// program exits with whatever the test binary exited with, so `go test` reads
// a run in the browser exactly as it reads a local one.
//
// [tgwasm]: https://pkg.go.dev/github.com/voilelab/toolgui/toolgui/tgwasm
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// runTimeout caps a run that never reports back: a browser that fails to
	// boot the binary would otherwise hang the build. It is longer than the
	// ten minutes `go test` gives a package by default, so a test that hangs
	// is cut off by its own timeout, with the goroutine dump that comes with
	// it, rather than by this.
	runTimeout = 15 * time.Minute

	// bootTimeout is how long the browser gets to load the page at all. It is
	// worth telling apart from a slow test run: nothing came back because
	// nothing started.
	bootTimeout = time.Minute
)

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "wasmtest:", err)
		os.Exit(1)
	}

	os.Exit(code)
}

// run returns the code the test binary exited with, so `go test` reads a run in
// the browser exactly as it reads a local one.
func run(args []string) (int, error) {
	if len(args) == 0 {
		return 0, errors.New("usage: wasmtest <test.wasm> [test flags]")
	}

	execJS, err := wasmExecJS()
	if err != nil {
		return 0, err
	}

	browser, err := findBrowser()
	if err != nil {
		return 0, err
	}

	// argv[0] is what the test binary reports as its own name; the rest are
	// the flags `go test` chose.
	argv, err := json.Marshal(append([]string{filepath.Base(args[0])}, args[1:]...))
	if err != nil {
		return 0, err
	}

	srv := &server{done: make(chan int, 1), loaded: make(chan struct{})}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.page)
	mux.HandleFunc("/worker.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprintf(w, workerJS, argv)
	})
	mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		http.ServeFile(w, r, execJS)
	})
	mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, args[0])
	})
	mux.HandleFunc("/out", srv.out)
	mux.HandleFunc("/exit", srv.exit)

	// 127.0.0.1 rather than a hostname: the file system API needs a secure
	// context, and loopback counts as one without a certificate.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()

	// Serve ends when the listener closes, which is on the way out of here.
	go func() { _ = http.Serve(ln, mux) }()

	return drive(browser, "http://"+ln.Addr().String(), srv)
}

// drive starts the browser on url and waits for the page to report how the
// test binary exited.
func drive(browser, url string, srv *server) (int, error) {
	// A profile of its own per run, so each one starts on an empty origin
	// private file system rather than on what the last run left.
	profile, err := os.MkdirTemp("", "toolgui-wasmtest-")
	if err != nil {
		return 0, err
	}
	defer func() { _ = os.RemoveAll(profile) }()

	cmd := exec.Command(browser,
		"--headless=new",
		"--no-sandbox",
		"--disable-gpu",
		// /dev/shm is small in a container, and Chrome falls over without
		// this when it is.
		"--disable-dev-shm-usage",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-search-engine-choice-screen",
		"--user-data-dir="+profile,
		url,
	)
	// Stdout and Stderr are left nil, which sends the browser's own logging to
	// /dev/null rather than through a pipe this would then have to drain. It
	// is noisy, and none of it is the test's.

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start %s: %w", browser, err)
	}

	// One Wait, on a goroutine of its own, so the result below can be waited
	// for at the same time. It reaps the browser after the kill.
	stopped := make(chan error, 1)
	go func() { stopped <- cmd.Wait() }()

	// The browser is never asked to quit: it has nothing left to do once the
	// result is in, and would otherwise sit there.
	defer func() { _ = cmd.Process.Kill() }()

	select {
	case <-srv.loaded:
	case err := <-stopped:
		return 0, fmt.Errorf("%s exited before it loaded the page: %v", browser, err)
	case <-time.After(bootTimeout):
		return 0, fmt.Errorf("%s did not load the page within %s", browser, bootTimeout)
	}

	select {
	case code := <-srv.done:
		return code, nil
	case err := <-stopped:
		return 0, fmt.Errorf("%s exited before the test did: %v", browser, err)
	case <-time.After(runTimeout):
		return 0, fmt.Errorf("no result after %s", runTimeout)
	}
}

// server takes the worker's output and its exit code back off the page.
type server struct {
	lock   sync.Mutex
	done   chan int
	loaded chan struct{}
	once   sync.Once
}

func (s *server) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s.once.Do(func() { close(s.loaded) })

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageHTML)
}

func (s *server) out(w http.ResponseWriter, r *http.Request) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// One line at a time from one page, but the writes are still serialized:
	// nothing should be able to interleave halfway through a test's output.
	s.lock.Lock()
	defer s.lock.Unlock()

	fmt.Println(string(bs))
}

func (s *server) exit(w http.ResponseWriter, r *http.Request) {
	code, err := strconv.Atoi(r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if bs, err := io.ReadAll(r.Body); err == nil && len(bs) > 0 {
		fmt.Fprintln(os.Stderr, "wasmtest:", string(bs))
	}

	select {
	case s.done <- code:
	default:
	}
}

// wasmExecJS finds the runtime shim of the toolchain in use. It's the
// toolchain's own, not a copy: a shim and the binary it boots go together.
func wasmExecJS() (string, error) {
	root := os.Getenv("GOROOT")
	if root == "" {
		out, err := exec.Command("go", "env", "GOROOT").Output()
		if err != nil {
			return "", fmt.Errorf("find GOROOT: %w", err)
		}

		root = strings.TrimSpace(string(out))
	}

	// lib/wasm since Go 1.24, misc/wasm before it.
	for _, dir := range []string{"lib", "misc"} {
		path := filepath.Join(root, dir, "wasm", "wasm_exec.js")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no wasm_exec.js under %s", root)
}

// browsers are the names a Chrome goes by, in the order they're tried. Set
// TOOLGUI_BROWSER to name one outright.
var browsers = []string{
	"google-chrome-stable",
	"google-chrome",
	"chromium-browser",
	"chromium",
	"chrome",
}

func findBrowser() (string, error) {
	if named := os.Getenv("TOOLGUI_BROWSER"); named != "" {
		return named, nil
	}

	for _, name := range browsers {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", errors.New("no Chrome or Chromium on PATH; set TOOLGUI_BROWSER")
}

// pageHTML boots the worker and relays what it says back to the server. The
// posts are chained so the output arrives in the order it was printed.
const pageHTML = `<!doctype html>
<meta charset="utf-8">
<title>toolgui wasmtest</title>
<script>
const worker = new Worker('worker.js')

let queue = Promise.resolve()
let done = false

function post(path, body) {
  queue = queue.then(() => fetch(path, { method: 'POST', body }))
  return queue
}

function finish(code, message) {
  if (done) return
  done = true
  post('exit?code=' + code, message || '')
}

worker.onmessage = (event) => {
  const message = event.data
  if (message.kind === 'out') {
    post('out', message.text)
  } else if (message.kind === 'exit') {
    finish(message.code, message.error)
  }
}

worker.onerror = (event) => {
  post('out', 'worker error: ' + event.message)
  finish(1)
}
</script>
`

// workerJS is where the test binary actually runs. Its %s is the argv, as
// JSON; the environment is left empty on purpose, because wasm_exec.js caps
// argv and environment together at a few kilobytes and a build machine's
// environment alone can be larger than that.
const workerJS = `const ARGV = %s

function post(kind, data) {
  self.postMessage(Object.assign({ kind }, data))
}

// Go's stdout and stderr both arrive as console.log, a line at a time.
for (const level of ['log', 'info', 'warn', 'error', 'debug']) {
  console[level] = (...args) => post('out', { text: args.map(String).join(' ') })
}

;(async () => {
  try {
    importScripts('wasm_exec.js')

    const go = new Go()
    go.argv = ARGV
    go.env = {}
    go.exit = (code) => post('exit', { code })

    const { instance } = await WebAssembly.instantiateStreaming(
      fetch('app.wasm'), go.importObject)

    await go.run(instance)
    post('exit', { code: 0 })
  } catch (e) {
    post('out', { text: 'wasmtest: ' + (e && e.stack || e) })
    post('exit', { code: 1, error: String(e) })
  }
})()
`
