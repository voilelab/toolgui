package tgframe

import (
	"errors"
	"testing"
)

// hostTestApp has two pages that do nothing.
func hostTestApp() *App {
	app := NewApp()
	blank := func(p *Params) error { return nil }
	app.AddPage("index", "Index", blank)
	app.AddPage("other", "Other", blank)
	return app
}

func TestSessionHostDoBeforeStart(t *testing.T) {
	h := NewSessionHost[int](hostTestApp(), func(any) error { return nil })

	err := h.Do(func(*State, map[string]int) error { return nil })
	if err != ErrNoSession {
		t.Errorf("Do = %v, want ErrNoSession", err)
	}

	if h.Session() != nil {
		t.Error("expect no session before a start")
	}
}

// TestSessionHostStartReplaces checks a second start drops the old state and
// its pending uploads, and cuts off the old session's sends.
func TestSessionHostStartReplaces(t *testing.T) {
	h := NewSessionHost[int](hostTestApp(), func(any) error { return nil })

	first, closeOld, err := h.Start("index", "")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	closeOld()

	var oldState *State
	_ = h.Do(func(state *State, uploads map[string]int) error {
		oldState = state
		uploads["u"] = 1
		_, err := state.SetFile("comp", "a.txt", []byte("hi"))
		return err
	})

	second, closeOld, err := h.Start("other", "")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// The old session is detached before it is closed.
	sendErr := first.sendPack(&ReadyPack{})

	closeOld()

	if !errors.Is(sendErr, ErrSessionReplaced) {
		t.Errorf("old session send = %v, want ErrSessionReplaced", sendErr)
	}

	if oldState.GetFile("comp") != nil {
		t.Error("expect the replaced state to be destroyed")
	}

	if h.Session() != second {
		t.Error("expect the new session to be current")
	}

	_ = h.Do(func(state *State, uploads map[string]int) error {
		if state == oldState {
			t.Error("expect a new state")
		}
		if len(uploads) != 0 {
			t.Errorf("%d uploads survived the start, want none", len(uploads))
		}
		return nil
	})
}

// TestSessionHostStartFails checks a failed start still closes the old
// session and leaves none.
func TestSessionHostStartFails(t *testing.T) {
	h := NewSessionHost[int](hostTestApp(), func(any) error { return nil })

	_, closeOld, err := h.Start("index", "")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	closeOld()

	session, closeOld, err := h.Start("nowhere", "")
	closeOld()

	if !errors.Is(err, ErrPageNotFound) {
		t.Errorf("Start = %v, want ErrPageNotFound", err)
	}

	if session != nil || h.Session() != nil {
		t.Error("expect no session after a failed start")
	}

	if err := h.Do(func(*State, map[string]int) error { return nil }); err != ErrNoSession {
		t.Errorf("Do = %v, want ErrNoSession", err)
	}
}

func TestSessionHostClose(t *testing.T) {
	h := NewSessionHost[int](hostTestApp(), func(any) error { return nil })

	if _, closeOld, err := h.Start("index", ""); err != nil {
		t.Fatalf("Start: %v", err)
	} else {
		closeOld()
	}

	h.Close()()

	if h.Session() != nil {
		t.Error("expect no session after Close")
	}

	// Closing with no session is a no-op.
	h.Close()()
}
