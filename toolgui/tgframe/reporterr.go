package tgframe

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
)

// InternalErrorMessage is all a client is told about a failure of the
// framework's own, or about a panic: that one happened. What it actually was
// goes to the log alone, under the error id the pack carries, so an operator
// given that id can find the line without the browser ever holding a function
// path, a file path or a panic value.
const InternalErrorMessage = "internal error"

// PageError marks an error a page meant its user to read -- one a page
// function returned, or one it reported with [Container.Fail]. Its message
// reaches the client unchanged; every other error is masked.
type PageError struct {
	Err error
}

// NewPageError wrap err as a page's own error. A nil err stays nil, so a page
// that returned nothing is not turned into a failure.
func NewPageError(err error) error {
	if err == nil {
		return nil
	}

	return &PageError{Err: err}
}

// Error return the message the page wrote.
func (e *PageError) Error() string {
	return e.Err.Error()
}

// Unwrap return the wrapped error, so errors.Is and errors.As still reach it.
func (e *PageError) Unwrap() error {
	return e.Err
}

// newErrorID return a short id naming one report. It only has to tell apart
// the reports an operator is reading at once, not to be unguessable.
func newErrorID() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(b[:])
}

// ReportError log err under msg and return the [ResultPack] telling the client
// about it. A [PageError] is a page talking to its own user, so its message
// goes out as it is; anything else is the framework's own and the client is
// given [InternalErrorMessage] and an error id that the log line carries too.
//
// A panicked error is never a page's own, whatever it wraps: what a panic
// carries is the process's, not something a page chose to show.
func ReportError(msg string, err error) *ResultPack {
	var pageErr *PageError
	if !errors.Is(err, ErrPanic) && errors.As(err, &pageErr) {
		slog.Error(msg, "error", err)

		return &ResultPack{
			Error:   pageErr.Error(),
			Success: false,
		}
	}

	id := newErrorID()
	slog.Error(msg, "error", err, "error_id", id)

	return &ResultPack{
		Error:   InternalErrorMessage,
		ErrorID: id,
		Success: false,
	}
}
