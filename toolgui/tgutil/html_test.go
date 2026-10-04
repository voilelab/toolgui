package tgutil

import "testing"

func TestInsertHead(t *testing.T) {
	cases := []struct {
		html, want string
		ok         bool
	}{
		{"<html><head></head></html>", "<html><head>\n<x></head></html>", true},
		{
			"<head>\n  <meta charset=\"utf-8\" />\n  <link rel=\"icon\" />",
			"<head>\n  <meta charset=\"utf-8\" />\n<x>\n  <link rel=\"icon\" />",
			true,
		},
		{"<HEAD lang=en><title>", "<HEAD lang=en>\n<x><title>", true},
		{"<body><header></header></body>", "<body><header></header></body>", false},
		{"<title>x</title>", "<title>x</title>", false},
	}

	for _, c := range cases {
		got, ok := InsertHead(c.html, "<x>")
		if got != c.want || ok != c.ok {
			t.Errorf("InsertHead(%q) = %q, %v; want %q, %v", c.html, got, ok, c.want, c.ok)
		}
	}
}
