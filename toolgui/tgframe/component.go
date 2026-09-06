package tgframe

import "fmt"

// Component is the interface of a component.
type Component interface {
	GetID() string
}

// keyed is what a component embedding [BaseComponent] gets for free. The
// container stamps a component's position through it; a component that does
// not embed BaseComponent is still a Component, it just carries no key of its
// own and the notify pack holds its position instead.
type keyed interface {
	GetKey() string
	setKey(key string)
}

// keyOf returns the position the container gave comp, or "" for a component
// that keeps none.
func keyOf(comp Component) string {
	if k, ok := comp.(keyed); ok {
		return k.GetKey()
	}
	return ""
}

var _ Component = &BaseComponent{}
var _ keyed = &BaseComponent{}

// BaseComponent stores the basic info of a component.
type BaseComponent struct {
	// Name is the typename of the component.
	Name string `json:"name"`

	// ID identifies the component across runs: it is the key its state is
	// stored under and the id it carries in the DOM. Components without state
	// leave it empty; when it is set it must be unique within a run.
	ID string `json:"id"`

	// key is the component's position, "<parent key>/<index>". It is what the
	// GUI client keys the node tree by, so it is stable across runs for a
	// component the page function writes in the same place. Assigned by
	// [Container.AddComponent] and sent on the notify pack, not in the props.
	key string
}

// GetID return component's ID.
func (c *BaseComponent) GetID() string {
	return c.ID
}

// SetID set component's ID.
func (c *BaseComponent) SetID(id string) {
	c.ID = fmt.Sprintf("%s_%s", c.Name, id)
}

// GetKey return component's key.
func (c *BaseComponent) GetKey() string {
	return c.key
}

func (c *BaseComponent) setKey(key string) {
	c.key = key
}
