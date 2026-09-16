package tgjson_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgjson"
)

type pack struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Map keys come out sorted, so one value is one byte sequence rather than one
// per run.
func TestMarshalIsDeterministic(t *testing.T) {
	in := map[string]int{"z": 1, "a": 2, "m": 3}

	bs, err := tgjson.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if got, want := string(bs), `{"a":2,"m":3,"z":1}`; got != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

// The v1 compatibility options are deliberately not set, so each of these
// reaches the caller as an error instead of passing quietly.
func TestUnmarshalRejects(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
	}{
		// v1 kept the last duplicate, which lets a sender smuggle a value
		// past anything that read the first one.
		{"duplicate object key", `{"name":"a","name":"b"}`},

		// 0xff is not valid UTF-8. v1 replaced it with U+FFFD, so what was
		// stored was not what arrived.
		{"invalid utf-8", "{\"name\":\"a\xffb\"}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got pack
			if err := tgjson.Unmarshal([]byte(tc.data), &got); err == nil {
				t.Errorf("Unmarshal(%q) = %+v, want an error", tc.data, got)
			}
		})
	}
}

// A name whose case does not match the tag no longer lands in the field. v1
// matched names case-insensitively, so "Name" and "NAME" both wrote Name and
// a sender could reach a field under a spelling the tag never gave it.
//
// It reads as an unknown member rather than an error: making it one needs
// RejectUnknownMembers, which would also reject the members a newer frontend
// sends to an older server, so the guarantee pinned here is that the value
// does not reach the field.
func TestUnmarshalDoesNotMatchNameCase(t *testing.T) {
	for _, name := range []string{"Name", "NAME", "nAmE"} {
		t.Run(name, func(t *testing.T) {
			got := pack{Name: "untouched"}
			if err := tgjson.Unmarshal([]byte(`{"`+name+`":"a"}`), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got.Name != "untouched" {
				t.Errorf("Name = %q, want it left alone", got.Name)
			}
		})
	}
}

// The exact-case name is still the one that lands.
func TestUnmarshalMatchesExactName(t *testing.T) {
	var got pack
	if err := tgjson.Unmarshal([]byte(`{"name":"a","value":2}`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Name != "a" || got.Value != 2 {
		t.Errorf("Unmarshal = %+v, want {a 2}", got)
	}
}

// MarshalWrite and UnmarshalRead are the same codec over a stream, which is
// what the handlers that write a response body and read one use.
func TestMarshalWriteUnmarshalRead(t *testing.T) {
	var buf bytes.Buffer
	if err := tgjson.MarshalWrite(&buf, pack{Name: "a", Value: 2}); err != nil {
		t.Fatalf("marshal write: %v", err)
	}

	var got pack
	if err := tgjson.UnmarshalRead(&buf, &got); err != nil {
		t.Fatalf("unmarshal read: %v", err)
	}

	if got.Name != "a" || got.Value != 2 {
		t.Errorf("round trip = %+v, want {a 2}", got)
	}
}

// A body with a second value after the first is an error, so a doubled or
// concatenated response does not read as just its first half.
func TestUnmarshalReadRejectsTrailingValue(t *testing.T) {
	var got pack
	err := tgjson.UnmarshalRead(strings.NewReader(`{"name":"a"}{"name":"b"}`), &got)
	if err == nil {
		t.Errorf("UnmarshalRead = %+v, want an error", got)
	}
}
