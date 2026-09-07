package main

import "testing"

func TestListenURL(t *testing.T) {
	cases := map[string]string{
		":3000":          "http://localhost:3000",
		"127.0.0.1:3000": "http://127.0.0.1:3000",
		"localhost:8080": "http://localhost:8080",
		"[::1]:3000":     "http://[::1]:3000",
		"not-an-address": "not-an-address",
	}

	for addr, want := range cases {
		got := listenURL(addr)
		if got != want {
			t.Errorf("listenURL(%q) = %q, want %q", addr, got, want)
		}
	}
}
