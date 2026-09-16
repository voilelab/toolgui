// Package tgjson is the only JSON codec toolgui uses. Every pack on the wire
// and every value read off it goes through here, so the options are set in one
// place and the same bytes cannot be encoded under two different semantics.
//
// It runs on [encoding/json/v2]. None of the v1 compatibility options are set:
// duplicate object keys, invalid UTF-8 and case-insensitive field names are
// errors, not something to be accepted quietly.
package tgjson

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io"
)

// options is what every call here encodes and decodes with. Deterministic
// makes map keys come out sorted, so a pack built from a map is one byte
// sequence rather than one per run, which is what lets a test compare output
// and a client cache a response.
var options = jsonv2.JoinOptions(
	jsonv2.Deterministic(true),
)

// Marshal encodes v as JSON.
func Marshal(v any) ([]byte, error) {
	return jsonv2.Marshal(v, options)
}

// Unmarshal decodes JSON data into v.
func Unmarshal(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v, options)
}

// MarshalWrite encodes v as JSON and writes it to w.
func MarshalWrite(w io.Writer, v any) error {
	return jsonv2.MarshalWrite(w, v, options)
}

// UnmarshalRead reads all of r and decodes it into v. Trailing content after
// the value is an error, so a truncated or doubled body does not pass.
func UnmarshalRead(r io.Reader, v any) error {
	return jsonv2.UnmarshalRead(r, v, options)
}

// MarshalEncode encodes v as the next value of enc. A [jsonv2.MarshalerTo]
// uses this to write its own representation without losing the options the
// surrounding encoder was given.
func MarshalEncode(enc *jsontext.Encoder, v any) error {
	return jsonv2.MarshalEncode(enc, v, options)
}
