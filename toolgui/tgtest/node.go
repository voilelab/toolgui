package tgtest

import (
	"bytes"
	"mime"
	"path/filepath"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

const formComponentName = "form_component"

// submitsForm are the components whose click sends the form they sit in, as
// on the web client. Any other click inside a form only waits for the submit.
var submitsForm = map[string]bool{
	"button_component": true,
}

// Node is a component as the last run drew it.
type Node struct {
	// Key is where the node sits in the tree, "<parent key>/<index>".
	Key string

	// Name is the component type, e.g. "button_component".
	Name string

	// ID is the component id, empty for a component without state.
	ID string

	// Props is the component as sent to the client.
	Props map[string]any

	Children []*Node

	page   *Page
	parent *Node
}

// File is one file to upload.
type File struct {
	Name string

	// Type is the mime type. Empty is guessed from Name.
	Type string

	Body []byte
}

// Prop returns the prop key, nil without one.
func (n *Node) Prop(key string) any {
	return n.Props[key]
}

// String returns the prop key as a string, "" unless it is one.
func (n *Node) String(key string) string {
	s, _ := n.Props[key].(string)
	return s
}

// Parent is the node n sits in, nil for a root container.
func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) walk(f func(n *Node)) {
	if n == nil {
		return
	}

	f(n)
	for _, c := range n.Children {
		c.walk(f)
	}
}

// queueKey is what a form's held events are kept under: the client's key
// for it, so a named form keeps them when it moves.
func (n *Node) queueKey() string {
	return reactKey(n.Key, n.Props)
}

// form is the form n sits in, nil outside one.
func (n *Node) form() *Node {
	for p := n.parent; p != nil; p = p.parent {
		if p.Name == formComponentName {
			return p
		}
	}

	return nil
}

// send applies event, or holds it for the form n sits in.
func (n *Node) send(event tgframe.Event) {
	n.page.t.Helper()

	if form := n.form(); form != nil {
		key := form.queueKey()
		n.page.forms[key] = append(n.page.forms[key], event)
		return
	}

	n.page.Send(event)
}

// usable fails the test unless a user could act on n: it has an id and is
// not disabled.
func (n *Node) usable(action string) {
	n.page.t.Helper()

	if n.ID == "" {
		n.page.t.Fatalf("tgtest: %s on %s at %s, which has no id",
			action, n.Name, n.Key)
	}

	if disabled, _ := n.Props["disabled"].(bool); disabled {
		n.page.t.Fatalf("tgtest: %s on disabled %s", action, n.ID)
	}
}

// Click clicks n. Inside a form a button click sends the form with it, and
// any other click waits for the submit.
func (n *Node) Click() {
	n.page.t.Helper()
	n.usable("click")

	if n.Name == "menu_component" {
		n.page.t.Fatalf("tgtest: click on menu %s, pick an item with Select", n.ID)
	}

	event := &tgframe.EventClick{ID: n.ID}

	form := n.form()
	if form == nil {
		n.page.Send(event)
		return
	}

	n.send(event)
	if submitsForm[n.Name] {
		form.Submit()
	}
}

// Input sets the value of n, e.g. a string for a textbox, a bool for a
// checkbox or a number for a number input. Inside a form it waits for the
// submit.
func (n *Node) Input(value any) {
	n.page.t.Helper()
	n.usable("input")

	n.send(&tgframe.EventInput{ID: n.ID, Value: n.wire(value)})
}

// wire passes v through json, so the page reads what it would off the wire:
// a float64 for any number, a map for a struct.
func (n *Node) wire(v any) any {
	n.page.t.Helper()

	bs, err := tgjson.Marshal(v)
	if err != nil {
		n.page.t.Fatalf("tgtest: encode value: %v", err)
	}

	var out any
	if err := tgjson.Unmarshal(bs, &out); err != nil {
		n.page.t.Fatalf("tgtest: decode value: %v", err)
	}

	return out
}

// Select picks the item at index i, 0-based: an option of a select, radio
// or select slider, or an item of a menu.
func (n *Node) Select(i int) {
	n.page.t.Helper()
	n.usable("select")

	switch n.Name {
	case "select_component":
		// 1-based on the wire, 0 being nothing picked.
		n.send(&tgframe.EventSelect{ID: n.ID, Value: i + 1})

	case "menu_component":
		// The click names the item, not the menu.
		items, _ := n.Props["items"].([]any)
		if i < 0 || i >= len(items) {
			n.page.t.Fatalf("tgtest: menu %s has no item %d", n.ID, i)
		}

		item, _ := items[i].(map[string]any)
		id, _ := item["id"].(string)
		event := &tgframe.EventClick{ID: id}

		form := n.form()
		n.send(event)
		if form != nil {
			form.Submit()
		}

	default:
		n.send(&tgframe.EventSelect{ID: n.ID, Value: i})
	}
}

// SelectMany picks the options at indexes of a multiselect.
func (n *Node) SelectMany(indexes ...int) {
	n.page.t.Helper()
	n.usable("select")

	if indexes == nil {
		indexes = []int{}
	}
	n.send(&tgframe.EventSelect{ID: n.ID, Values: indexes})
}

// SelectKeys picks the rows of keys in a DataFrame with row keys.
func (n *Node) SelectKeys(keys ...string) {
	n.page.t.Helper()
	n.usable("select")

	if keys == nil {
		keys = []string{}
	}
	n.send(&tgframe.EventSelect{ID: n.ID, Keys: keys})
}

// Upload picks a file of name and body in a fileupload.
func (n *Node) Upload(name string, body []byte) {
	n.page.t.Helper()
	n.UploadFiles(File{Name: name, Body: body})
}

// UploadFiles picks files in a fileupload; a single-file one takes exactly
// one. Like the browser, the files are stored before the pick is sent.
func (n *Node) UploadFiles(files ...File) {
	n.page.t.Helper()
	n.usable("upload")

	t, state := n.page.t, n.page.state
	multiple, _ := n.Props["multiple"].(bool)

	if !multiple && len(files) != 1 {
		t.Fatalf("tgtest: %d files to single-file %s", len(files), n.ID)
	}

	metas := make([]map[string]any, len(files))
	for i, f := range files {
		key := n.ID
		if multiple {
			key = tgframe.FileKey(n.ID, i)
		}

		// The same check the upload endpoint makes.
		if !state.HasFileKey(key) {
			t.Fatalf("tgtest: %s does not take uploads", key)
		}

		if _, err := state.WriteFile(key, f.Name, bytes.NewReader(f.Body)); err != nil {
			t.Fatalf("tgtest: store upload %s: %v", f.Name, err)
		}

		typ := f.Type
		if typ == "" {
			typ = mime.TypeByExtension(filepath.Ext(f.Name))
		}

		metas[i] = map[string]any{
			"name": f.Name,
			"type": typ,
			"size": len(f.Body),
		}
	}

	if multiple {
		n.send(&tgframe.EventInput{ID: n.ID, Value: n.wire(metas)})
		return
	}

	n.send(&tgframe.EventInput{ID: n.ID, Value: n.wire(metas[0])})
}

// Submit sends form n with what was written in it since the last submit,
// like its submit button.
func (n *Node) Submit() {
	n.page.t.Helper()

	if n.Name != formComponentName {
		n.page.t.Fatalf("tgtest: submit on %s at %s, which is not a form",
			n.Name, n.Key)
	}

	events := n.page.forms[n.queueKey()]
	delete(n.page.forms, n.queueKey())

	if events == nil {
		events = []tgframe.Event{}
	}
	n.page.Send(&tgframe.EventForm{Events: events})
}
