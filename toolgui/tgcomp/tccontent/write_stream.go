package tccontent

import (
	"iter"
	"strings"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

const defaultWriteStreamInterval = 50 * time.Millisecond

// WriteStreamConf is the configuration for the WriteStream component.
//
//tgcomp:export
type WriteStreamConf struct {
	tgframe.Base

	// Interval is the least time between two sends of the text. Default is
	// 50ms.
	Interval time.Duration
}

// WriteStream writes the chunks of seq into one Markdown as they arrive, and
// returns the whole text. The first error of seq stops it, and is returned
// with the text before it.
//
// The growing text is sent at most once per Conf.Interval, not once per chunk.
//
// seq runs on a goroutine of its own and must not draw components. WriteStream
// waits for it to return, so build it on [tgframe.Params.Context]: a cut run
// then ends it at once.
//
//tgcomp:export
func WriteStream(c *tgframe.Container, seq iter.Seq2[string, error],
	conf ...*WriteStreamConf) (string, error) {
	cf := tgframe.OneConf("WriteStream", conf)

	interval := cf.Interval
	if interval <= 0 {
		interval = defaultWriteStreamInterval
	}

	comp := newMarkdownComponent("")
	tgframe.SetConfIDIn(c, comp, cf)
	c.AddComponent(comp)

	chunks := pullChunks(seq)
	defer chunks.stop()

	var sb strings.Builder
	sent := 0
	flush := func() {
		if sb.Len() == sent {
			return
		}

		comp.Markdown = sb.String()
		sent = sb.Len()
		c.SendNotifyPack(tgframe.NewNotifyPackUpdate(comp))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case ch, ok := <-chunks.c:
			switch {
			case !ok:
				flush()
				return sb.String(), nil
			case ch.panic != nil:
				panic(ch.panic)
			case ch.err != nil:
				flush()
				return sb.String(), ch.err
			}

			sb.WriteString(ch.text)
		case <-ticker.C:
			flush()
		}
	}
}

type streamChunk struct {
	text  string
	err   error
	panic any
}

// chunkPuller runs a seq on its own goroutine, so a chunk can be waited on
// together with a timer.
type chunkPuller struct {
	c      chan streamChunk
	done   chan struct{}
	exited chan struct{}
}

func pullChunks(seq iter.Seq2[string, error]) *chunkPuller {
	p := &chunkPuller{
		c:      make(chan streamChunk),
		done:   make(chan struct{}),
		exited: make(chan struct{}),
	}

	go func() {
		defer close(p.exited)
		defer close(p.c)
		// Handed to the page goroutine, where a run's panics are caught.
		defer func() {
			if r := recover(); r != nil {
				p.send(streamChunk{panic: r})
			}
		}()

		for text, err := range seq {
			if !p.send(streamChunk{text: text, err: err}) || err != nil {
				return
			}
		}
	}()

	return p
}

func (p *chunkPuller) send(ch streamChunk) bool {
	select {
	case p.c <- ch:
		return true
	case <-p.done:
		return false
	}
}

// stop ends the seq at its next yield and waits for it to return.
func (p *chunkPuller) stop() {
	close(p.done)
	<-p.exited
}
