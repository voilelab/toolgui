package tgframe

import "fmt"

var _ Component = &Container{}
var ContainerComponentName = "container_component"

// Container contain list of components.
type Container struct {
	*BaseComponent

	SendNotifyPack SendNotifyPackFunc `json:"-"`

	// counter is the index the next component added here gets. Containers are
	// rebuilt on every run, so it starts at 0 each time and a component keeps
	// its index as long as the page function writes it in the same place.
	counter int

	// run is shared by every container of a run. Nil outside a run.
	run *runState
}

func NewContainer(id string, notifyComp SendNotifyPackFunc) *Container {
	return &Container{
		BaseComponent: &BaseComponent{
			Name: ContainerComponentName,
			ID:   containerID(id),

			// A root container is never sent, so nothing assigns it a key.
			// Its own id is the root of every key below it.
			key: containerID(id),
		},
		SendNotifyPack: notifyComp,
	}
}

func containerID(id string) string {
	return fmt.Sprintf("%s_%s", ContainerComponentName, id)
}

func (c *Container) AddComponent(comp Component) Component {
	idx := c.counter
	c.counter++

	key := fmt.Sprintf("%s/%d", c.key, idx)
	if k, ok := comp.(keyed); ok {
		k.setKey(key)
	}

	if c.run != nil {
		c.run.registerID(comp)
	}

	c.SendNotifyPack(NewNotifyPackCreate(c.key, idx, key, comp))
	return comp
}

func (c *Container) AddContainer(id string) *Container {
	newContainer := NewContainer(id, c.SendNotifyPack)
	newContainer.run = c.run
	c.AddComponent(newContainer)
	return newContainer
}

// AddContainerTo creates the idx-th container inside comp, a component this
// container has already added. Layout components use it for the containers
// they own, so that those containers sit under the component in the node tree.
//
// The container is given an id derived from comp's, or none when comp has
// none: a component that does not claim an identity does not hand one out.
func (c *Container) AddContainerTo(comp Component, suffix string, idx int) *Container {
	inner := &Container{
		BaseComponent: &BaseComponent{
			Name: ContainerComponentName,
			key:  fmt.Sprintf("%s/%d", keyOf(comp), idx),
		},
		SendNotifyPack: c.SendNotifyPack,
		run:            c.run,
	}

	if comp.GetID() != "" {
		inner.ID = comp.GetID() + "_" + suffix
	}

	if c.run != nil {
		c.run.registerID(inner)
	}

	c.SendNotifyPack(NewNotifyPackCreate(keyOf(comp), idx, inner.key, inner))
	return inner
}

// With is a helper function to add a component to the container.
// Example:
//
//	container.With(func(c *Container) {
//		Button(c, "button", "Click me"))
//	})
func (c *Container) With(f func(c *Container)) {
	f(c)
}
