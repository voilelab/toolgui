package tgframe

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestParseQuery(t *testing.T) {
	q, err := ParseQuery("group=a&name=x+y")
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}

	if q.Get("group") != "a" || q.Get("name") != "x y" {
		t.Errorf("query = %v", q)
	}
}

func TestParseQueryErrors(t *testing.T) {
	big := "x=" + strings.Repeat("a", MaxQuerySize)
	if _, err := ParseQuery(big); !errors.Is(err, ErrQueryTooLarge) {
		t.Errorf("oversized: err = %v", err)
	}

	_, err := ParseQuery("x=%zzsecret")
	if !errors.Is(err, ErrInvalidQuery) {
		t.Errorf("invalid: err = %v", err)
	}

	if err != nil && strings.Contains(err.Error(), "secret") {
		t.Error("the error carries the query")
	}
}

func TestNewSessionOversizedQuery(t *testing.T) {
	app := NewApp()
	app.AddPage(testPageName, "Test", func(p *Params) error { return nil })

	query := url.Values{"x": {strings.Repeat("a", MaxQuerySize)}}
	_, err := NewSession(app, testPageName, query, NewState(),
		func(any) error { return nil })
	if !errors.Is(err, ErrQueryTooLarge) {
		t.Errorf("err = %v, want ErrQueryTooLarge", err)
	}
}
