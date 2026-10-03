package tccontent

import (
	"errors"
	"iter"
	"sync"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// streamPacks runs WriteStream and returns the props of every pack it sent, in
// order, and its result.
func streamPacks(t *testing.T, seq iter.Seq2[string, error],
	conf ...*WriteStreamConf) ([]map[string]any, string, error) {
	t.Helper()

	var mu sync.Mutex
	var props []map[string]any
	c := tgframe.NewContainer("test", tgframe.NewState(), func(pack tgframe.NotifyPack) {
		bs, err := tgjson.Marshal(pack)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var out struct {
			Component map[string]any `json:"component"`
		}
		if err := tgjson.Unmarshal(bs, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		mu.Lock()
		props = append(props, out.Component)
		mu.Unlock()
	})

	text, err := WriteStream(c, seq, conf...)
	return props, text, err
}

func chunks(texts ...string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, s := range texts {
			if !yield(s, nil) {
				return
			}
		}
	}
}

func TestWriteStreamText(t *testing.T) {
	props, text, err := streamPacks(t, chunks("Hello", ", ", "**World**"),
		&WriteStreamConf{Base: tgframe.Base{ID: "reply"}, Interval: time.Hour})
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	if text != "Hello, **World**" {
		t.Errorf("text = %q", text)
	}

	// An hour-long interval leaves the create and the final flush only.
	if len(props) != 2 {
		t.Fatalf("got %d packs, want 2", len(props))
	}

	if props[0]["name"] != markdownComponentName || props[0]["text"] != "" {
		t.Errorf("create = %v, want an empty markdown", props[0])
	}

	if props[1]["text"] != text || props[1]["id"] != "markdown_component_reply" {
		t.Errorf("last = %v, want text %q id markdown_component_reply", props[1], text)
	}
}

func TestWriteStreamEmpty(t *testing.T) {
	props, text, err := streamPacks(t, chunks())
	if text != "" || err != nil {
		t.Errorf("got (%q, %v), want empty", text, err)
	}

	if len(props) != 1 {
		t.Errorf("got %d packs, want only the create", len(props))
	}
}

func TestWriteStreamFlushesWhileStreaming(t *testing.T) {
	release := make(chan struct{})
	var seen []map[string]any

	seq := func(yield func(string, error) bool) {
		if !yield("first", nil) {
			return
		}
		<-release
		yield(" second", nil)
	}

	var mu sync.Mutex
	c := tgframe.NewContainer("test", tgframe.NewState(), func(pack tgframe.NotifyPack) {
		bs, _ := tgjson.Marshal(pack)
		var out struct {
			Component map[string]any `json:"component"`
		}
		_ = tgjson.Unmarshal(bs, &out)

		mu.Lock()
		seen = append(seen, out.Component)
		mu.Unlock()

		// The first chunk reached the client before the seq went on.
		if out.Component["text"] == "first" {
			close(release)
		}
	})

	text, err := WriteStream(c, seq, &WriteStreamConf{Interval: time.Millisecond})
	if err != nil || text != "first second" {
		t.Fatalf("got (%q, %v)", text, err)
	}

	if last := seen[len(seen)-1]["text"]; last != "first second" {
		t.Errorf("last text = %v", last)
	}
}

func TestWriteStreamError(t *testing.T) {
	boom := errors.New("boom")
	seq := func(yield func(string, error) bool) {
		if !yield("partial", nil) {
			return
		}
		yield("ignored", boom)
	}

	props, text, err := streamPacks(t, seq, &WriteStreamConf{Interval: time.Hour})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want boom", err)
	}

	if text != "partial" || props[len(props)-1]["text"] != "partial" {
		t.Errorf("text = %q, last = %v, want partial", text, props[len(props)-1])
	}
}

func TestWriteStreamPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != "boom" {
			t.Errorf("recover = %v, want boom", r)
		}
	}()

	streamPacks(t, func(yield func(string, error) bool) { panic("boom") })
	t.Error("WriteStream returned, want a panic")
}

// A run cut at a send stops the seq, and WriteStream waits for it to return.
func TestWriteStreamInterrupt(t *testing.T) {
	stopped := false
	seq := func(yield func(string, error) bool) {
		for yield("x", nil) {
			time.Sleep(time.Millisecond)
		}
		stopped = true
	}

	c := tgframe.NewContainer("test", tgframe.NewState(), func(pack tgframe.NotifyPack) {
		if pack.GetType() == tgframe.NotifyTypeUpdate {
			panic(tgframe.ErrUpdateInterrupt)
		}
	})

	func() {
		defer func() {
			if r := recover(); r != tgframe.ErrUpdateInterrupt {
				t.Errorf("recover = %v, want ErrUpdateInterrupt", r)
			}
		}()
		_, _ = WriteStream(c, seq, &WriteStreamConf{Interval: time.Millisecond})
	}()

	if !stopped {
		t.Error("seq still running after WriteStream unwound")
	}
}
