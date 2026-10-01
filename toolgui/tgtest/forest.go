package tgtest

import "github.com/voilelab/toolgui/toolgui/tgframe"

// notifyPack is a notify pack as it arrives on the wire.
type notifyPack struct {
	Type      int            `json:"type"`
	ParentKey string         `json:"parent_key"`
	Index     int            `json:"index"`
	Key       string         `json:"key"`
	Component map[string]any `json:"component"`
}

// fnode is a node of the tree the packs build, before it is snapshotted.
type fnode struct {
	key       string
	parentKey string
	props     map[string]any
	children  []*fnode
	runID     int
}

func (n *fnode) name() string {
	s, _ := n.props["name"].(string)
	return s
}

// reactKey is what the web client keys n by: its id, else position and type.
func (n *fnode) reactKey() string {
	return reactKey(n.key, n.props)
}

func reactKey(key string, props map[string]any) string {
	if id, _ := props["id"].(string); id != "" {
		return id
	}

	name, _ := props["name"].(string)
	return key + ":" + name
}

// forest mirrors the web client's node tree (toolgui-web Nodes.ts), so a
// page is seen the way a browser would show it.
type forest struct {
	nodes map[string]*fnode
	roots []string
	runID int
}

func newForest(roots ...string) *forest {
	f := &forest{nodes: map[string]*fnode{}, roots: roots}
	for _, id := range roots {
		f.nodes[id] = &fnode{
			key: id,
			props: map[string]any{
				"name": tgframe.ContainerComponentName,
				"id":   id,
			},
		}
	}

	return f
}

func (f *forest) beginRun() {
	f.runID++
	for _, id := range f.roots {
		f.nodes[id].runID = f.runID
	}
}

func (f *forest) apply(pack *notifyPack) {
	switch pack.Type {
	case tgframe.NotifyTypeCreate:
		f.create(pack.Key, pack.ParentKey, pack.Index, pack.Component)
	case tgframe.NotifyTypeUpdate:
		if n := f.nodes[pack.Key]; n != nil {
			n.props = pack.Component
		}
	case tgframe.NotifyTypeDelete:
		f.remove(pack.Key)
	}
}

func (f *forest) create(key, parentKey string, index int, props map[string]any) {
	parent := f.nodes[parentKey]
	if parent == nil || index < 0 {
		return
	}

	// A different component type at this position doesn't inherit children.
	node := f.nodes[key]
	if node == nil || node.name() != props["name"] {
		node = &fnode{key: key}
	}

	node.props = props
	node.runID = f.runID
	node.parentKey = parentKey
	f.nodes[key] = node

	// A named node that moved leaves its old instance behind until endRun,
	// which a failed run skips. Retire it now, as the web client does.
	rk := node.reactKey()
	for staleKey, stale := range f.nodes {
		if staleKey != key && stale.runID != f.runID && stale.reactKey() == rk {
			f.remove(staleKey)
		}
	}

	for len(parent.children) <= index {
		parent.children = append(parent.children, nil)
	}
	parent.children[index] = node
}

func (f *forest) remove(key string) {
	node := f.nodes[key]
	if node == nil {
		return
	}

	if parent := f.nodes[node.parentKey]; parent != nil {
		kept := parent.children[:0]
		for _, c := range parent.children {
			if c != node {
				kept = append(kept, c)
			}
		}
		parent.children = kept
	}
	delete(f.nodes, key)

	for _, c := range node.children {
		if c != nil {
			f.remove(c.key)
		}
	}
}

// endRun drops what the run didn't send. A failed run stopped partway, so
// the tree is left alone.
func (f *forest) endRun(success bool) {
	if !success {
		return
	}

	for key, n := range f.nodes {
		if n.runID != f.runID {
			delete(f.nodes, key)
		}
	}

	for _, n := range f.nodes {
		kept := n.children[:0]
		for _, c := range n.children {
			if c != nil && c.runID == f.runID {
				kept = append(kept, c)
			}
		}
		n.children = kept
	}
}

// snapshot copies the tree under key, so what a test holds stays put when
// the next run rewrites the forest.
func (f *forest) snapshot(p *Page, key string, parent *Node) *Node {
	n := f.nodes[key]
	if n == nil {
		return nil
	}

	props := make(map[string]any, len(n.props))
	for k, v := range n.props {
		props[k] = v
	}

	node := &Node{
		Key:    n.key,
		Props:  props,
		page:   p,
		parent: parent,
	}
	node.Name, _ = props["name"].(string)
	node.ID, _ = props["id"].(string)

	for _, c := range n.children {
		if c == nil {
			continue
		}

		if child := f.snapshot(p, c.key, node); child != nil {
			node.Children = append(node.Children, child)
		}
	}

	return node
}
