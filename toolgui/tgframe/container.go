package tgframe

import "fmt"

var _ Component = &Container{}
var ContainerComponentName = "container_component"

// Container contain list of components.
type Container struct {
	*BaseComponent

	SendNotifyPack SendNotifyPackFunc `json:"-"`

	// State is the session state a component added here reads and writes,
	// instead of being handed a State of its own. [App.Run] gives every
	// container of a run the same one; a container built directly through
	// [NewContainer] carries whatever its caller passed, which may be nil.
	State *State `json:"-"`

	// counter is the index the next component added here gets. Containers are
	// rebuilt on every run, so it starts at 0 each time and a component keeps
	// its index as long as the page function writes it in the same place.
	counter int

	// run is shared by every container of a run. Nil outside a run.
	run *runState

	// track collects what is written here and below, for the [Slot] that has
	// to be able to take it all back off the screen. Nil outside a slot.
	track *tracker
}

func NewContainer(id string, state *State, notifyComp SendNotifyPackFunc) *Container {
	return &Container{
		BaseComponent: &BaseComponent{
			Name: ContainerComponentName,
			ID:   containerID(id),

			// A root container is never sent, so nothing assigns it a key.
			// Its own id is the root of every key below it.
			key: containerID(id),
		},
		SendNotifyPack: notifyComp,
		State:          state,
	}
}

func containerID(id string) string {
	return fmt.Sprintf("%s_%s", ContainerComponentName, id)
}

// innerKey is where the idx-th container of comp sits in the node tree.
func innerKey(comp Component, idx int) string {
	return fmt.Sprintf("%s/%d", keyOf(comp), idx)
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
	c.track.add(comp)

	c.SendNotifyPack(NewNotifyPackCreate(c.key, idx, key, comp))
	return comp
}

func (c *Container) AddContainer(id string) *Container {
	newContainer := NewContainer(id, c.State, c.SendNotifyPack)
	newContainer.run = c.run
	newContainer.track = c.track
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
	inner := c.innerContainer(comp, suffix, idx)

	if c.run != nil {
		c.run.registerID(inner)
	}
	c.track.add(inner)

	c.SendNotifyPack(NewNotifyPackCreate(keyOf(comp), idx, inner.key, inner))
	return inner
}

// innerContainer builds the idx-th container of comp without sending it, so
// that [Container.AddContainerTo] and [Slot] agree on what one looks like.
func (c *Container) innerContainer(comp Component, suffix string, idx int) *Container {
	inner := &Container{
		BaseComponent: &BaseComponent{
			Name: ContainerComponentName,
			key:  innerKey(comp, idx),
		},
		SendNotifyPack: c.SendNotifyPack,
		State:          c.State,
		run:            c.run,
		track:          c.track,
	}

	if comp.GetID() != "" {
		inner.ID = comp.GetID() + "_" + suffix
	}

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

// RemoveComponent takes comp off the screen and gives its id back, so this run
// may claim it again and [App.Run] drops the state under an id nothing claims.
// It is [Container.AddComponent]'s counterpart, for a component that takes
// itself down before the run ends; [Slot.Clear] is the same move for a slot.
//
// Only comp's own id is given back, not those of anything it added below
// itself.
func (c *Container) RemoveComponent(comp Component) {
	c.SendNotifyPack(NewNotifyPackDelete(comp))

	if c.run != nil {
		c.run.unregisterID(comp)
	}
}
