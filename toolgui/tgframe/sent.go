package tgframe

import (
	"bytes"
	"encoding/json/jsontext"
	"hash/maphash"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// sentNode is what the client holds at a key, as far as the session knows.
type sentNode struct {
	// sum and size identify the component bytes last sent here. ok is false
	// after an update pack, whose props the cache did not hash.
	sum  uint64
	size int
	ok   bool

	id string

	// seq is the run that last sent the node, create or keep.
	seq uint64
}

// sentCache mirrors the client's node tree (toolgui-web Nodes.ts) by key, so a
// create whose bytes the client already holds goes out as a keep pack.
//
// It only has to never claim a node the client lacks: forgetting one costs a
// full create, nothing more. Hence a delete drops the whole key prefix and an
// update drops the hash.
type sentCache struct {
	seed  maphash.Seed
	nodes map[string]*sentNode

	// byID is the keys holding each non-empty id, for the stale node the
	// client drops when a named component moves.
	byID map[string]map[string]bool

	// scanned is the run each id was last scanned in. After a scan every key
	// holding the id is sent this run, and a key gains an id only by being
	// sent, so one scan per id per run is enough.
	scanned map[string]uint64

	seq uint64

	// buf is reused across marshals, so a keep allocates nothing for the
	// component bytes.
	buf bytes.Buffer
}

func newSentCache() *sentCache {
	return &sentCache{
		seed:  maphash.MakeSeed(),
		nodes: map[string]*sentNode{},
		byID:  map[string]map[string]bool{},

		scanned: map[string]uint64{},
	}
}

// beginRun mirrors the client's beginRun on a ready pack.
func (c *sentCache) beginRun() {
	c.seq++
}

// endRun mirrors the client's endRun on a successful result: whatever the run
// did not send is gone.
func (c *sentCache) endRun() {
	for key, n := range c.nodes {
		if n.seq != c.seq {
			c.drop(key)
		}
	}
}

// filter returns what to send for pack: a keep pack for a create the client
// already holds, a create carrying its marshaled bytes otherwise, or pack
// itself.
func (c *sentCache) filter(pack NotifyPack) (any, error) {
	switch p := pack.(type) {
	case *notifyPackCreate:
		c.buf.Reset()
		err := tgjson.MarshalWrite(&c.buf, p.Component)
		if err != nil {
			return nil, tgutil.Errorf("%w", err)
		}

		// MarshalWrite ends with a newline, which is no part of the value.
		bs := bytes.TrimSuffix(c.buf.Bytes(), []byte("\n"))
		if c.create(p.Key, p.Component.GetID(), bs) {
			return NewNotifyPackKeep(p.ParentKey, p.Index, p.Key), nil
		}

		return &notifyPackCreateRaw{
			notifyPackBase: p.notifyPackBase,
			ParentKey:      p.ParentKey,
			Index:          p.Index,
			Key:            p.Key,
			Component:      jsontext.Value(bytes.Clone(bs)),
		}, nil

	case *notifyPackUpdate:
		if n := c.nodes[p.Key]; n != nil {
			n.ok = false
		}

	case *notifyPackDelete:
		c.remove(p.Key)
	}

	return pack, nil
}

// create records bs sent at key and reports whether the client already holds
// exactly that.
func (c *sentCache) create(key, id string, bs []byte) bool {
	sum := maphash.Bytes(c.seed, bs)

	n := c.nodes[key]
	same := n != nil && n.ok && n.sum == sum && n.size == len(bs)
	if n == nil {
		n = &sentNode{}
		c.nodes[key] = n
	}

	if n.id != id {
		c.unindex(key, n.id)
		n.id = id
		c.index(key, id)
	}

	n.sum, n.size, n.ok = sum, len(bs), true
	n.seq = c.seq

	// The client retires a node of the same id this run has not sent, with
	// its subtree. A node without an id is keyed by its position, so it
	// never collides with another key.
	if id != "" && c.scanned[id] != c.seq {
		c.scanned[id] = c.seq
		for other := range c.byID[id] {
			if other != key && c.nodes[other].seq != c.seq {
				c.remove(other)
			}
		}
	}

	return same
}

// remove drops key and everything under it. Every child key extends its
// parent's, so this covers the client's subtree and more.
func (c *sentCache) remove(key string) {
	prefix := key + "/"
	for k := range c.nodes {
		if k == key || strings.HasPrefix(k, prefix) {
			c.drop(k)
		}
	}
}

func (c *sentCache) drop(key string) {
	if n := c.nodes[key]; n != nil {
		c.unindex(key, n.id)
		delete(c.nodes, key)
	}
}

func (c *sentCache) index(key, id string) {
	if id == "" {
		return
	}

	if c.byID[id] == nil {
		c.byID[id] = map[string]bool{}
	}
	c.byID[id][key] = true
}

func (c *sentCache) unindex(key, id string) {
	if id == "" {
		return
	}

	delete(c.byID[id], key)
	if len(c.byID[id]) == 0 {
		delete(c.byID, id)
		delete(c.scanned, id)
	}
}
