package tgframe

// tracker collects the components written into a [Slot], so that clearing the
// slot can give their ids back. A slot inside a slot chains to the outer one:
// clearing the outer takes what the inner holds with it.
type tracker struct {
	outer *tracker
	comps []Component
}

// add records comp here and in every slot this one sits in. A nil tracker is
// a container outside any slot, where nothing has to be taken back.
func (t *tracker) add(comp Component) {
	for cur := t; cur != nil; cur = cur.outer {
		cur.comps = append(cur.comps, comp)
	}
}

// Slot is a place in the page that can be written and written over. What it
// holds is one container's worth of components, and writing it again first
// takes the previous contents off the screen rather than adding to them, so a
// page function can show "loading…" and then replace it with the result.
//
// It is what [github.com/voilelab/toolgui/toolgui/tgcomp.Empty] hands out.
type Slot struct {
	// owner is the container that sent comp, the component the slot's
	// contents hang under. idx and suffix are where under it they hang.
	owner  *Container
	comp   Component
	idx    int
	suffix string

	// key is fixed for the life of the slot, so the client keeps one node for
	// it however many times it is rewritten.
	key string

	// inner is the container the current contents went into, nil while the
	// slot is empty, and track is what they wrote.
	inner *Container
	track *tracker
}

// AddSlotTo creates the idx-th child of comp as a slot: a container this
// container can take back off the screen and write again. Ids and keys are
// [Container.AddContainerTo]'s, so a slot is that container plus the ability
// to clear it.
//
// The slot starts empty, whatever the previous run left in its place.
func (c *Container) AddSlotTo(comp Component, suffix string, idx int) *Slot {
	s := &Slot{
		owner:  c,
		comp:   comp,
		idx:    idx,
		suffix: suffix,
		key:    innerKey(comp, idx),
	}

	c.SendNotifyPack(NewNotifyPackDeleteKey(s.key))
	return s
}

// Clear takes what the slot holds off the screen, and gives back the ids and
// the state that went with it. Clearing an empty slot does nothing.
func (s *Slot) Clear() {
	if s.inner == nil {
		return
	}

	s.owner.SendNotifyPack(NewNotifyPackDeleteKey(s.key))

	if run := s.owner.run; run != nil {
		run.unregisterID(s.inner)
		for _, comp := range s.track.comps {
			run.unregisterID(comp)
		}
	}

	s.inner = nil
	s.track = nil
}

// With writes what f adds into the slot, over whatever it held. The container
// f is handed lives until the next [Slot.With] or [Slot.Clear]; writing into
// it after that puts components under a node the client no longer has.
func (s *Slot) With(f func(c *Container)) {
	s.Clear()

	inner := s.owner.innerContainer(s.comp, s.suffix, s.idx)
	s.track = &tracker{outer: s.owner.track}
	inner.track = s.track

	if s.owner.run != nil {
		s.owner.run.registerID(inner)
	}
	s.inner = inner

	s.owner.SendNotifyPack(
		NewNotifyPackCreate(keyOf(s.comp), s.idx, inner.key, inner))

	f(inner)
}
