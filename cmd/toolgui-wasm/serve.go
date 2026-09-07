package main

import (
	"flag"
	"log"
	"mime"
	"net"
	"net/http"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

func runServe(args []string) error {
	addr := ""
	out, pkg, err := parseBuildFlags("serve", args, func(flags *flag.FlagSet) {
		flags.StringVar(&addr, "addr", ":3000", "address to listen on")
	})
	if err != nil {
		return err
	}

	err = build(out, pkg)
	if err != nil {
		return err
	}

	// So the browser compiles the binary while it downloads.
	err = mime.AddExtensionType(".wasm", "application/wasm")
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	log.Printf("serving %s on %s", out, listenURL(addr))

	err = http.ListenAndServe(addr, http.FileServer(http.Dir(out)))
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// listenURL turn a listen address into one that can be opened. An address
// with no host listens on every interface, and localhost is the one the
// person who ran the command is at.
func listenURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	if host == "" {
		host = "localhost"
	}

	return "http://" + net.JoinHostPort(host, port)
}
