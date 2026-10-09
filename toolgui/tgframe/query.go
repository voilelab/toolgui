package tgframe

import (
	"errors"
	"net/url"
)

// MaxQuerySize caps a page query, encoded. A query is anyone's to build, and
// every run gets a copy of it.
const MaxQuerySize = 8 * 1024

// ErrQueryTooLarge is the error that a page query is over [MaxQuerySize].
var ErrQueryTooLarge = errors.New("page query too large")

// ErrInvalidQuery is the error that a page query does not parse.
var ErrInvalidQuery = errors.New("invalid page query")

// ParseQuery parses a raw page query, `group=a&name=x`, for [NewSession].
// The errors carry no part of the query: a value can be data.
func ParseQuery(raw string) (url.Values, error) {
	if len(raw) > MaxQuerySize {
		return nil, ErrQueryTooLarge
	}

	query, err := url.ParseQuery(raw)
	if err != nil {
		return nil, ErrInvalidQuery
	}

	return query, nil
}

// QueryPack tells the client to replace the page query in its address bar,
// without a reload or a history entry. See [Params.ReplaceQuery].
type QueryPack struct {
	// ReplaceQuery is never nil, so the client always reads an object.
	ReplaceQuery url.Values `json:"replace_query"`
}

// Navigation is a page of this app and the query to open it with.
type Navigation struct {
	Page string `json:"page"`

	// Query is never nil, so the client always reads an object.
	Query url.Values `json:"query"`
}

// NavigatePack tells the client to open another page, as a click on a
// tgcomp.PageLink would: a new session and a history entry. See
// [Params.Navigate].
type NavigatePack struct {
	Navigate *Navigation `json:"navigate"`
}

// checkQuery reports whether query fits in [MaxQuerySize].
func checkQuery(query url.Values) error {
	if len(query.Encode()) > MaxQuerySize {
		return ErrQueryTooLarge
	}

	return nil
}

// cloneQuery copies query deep, so a page writing to its copy cannot reach the
// session's. A nil query is an empty one.
func cloneQuery(query url.Values) url.Values {
	clone := make(url.Values, len(query))
	for k, vs := range query {
		clone[k] = append([]string(nil), vs...)
	}

	return clone
}
