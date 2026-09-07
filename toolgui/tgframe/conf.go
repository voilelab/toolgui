package tgframe

// Base is embedded in every component's Conf, and is what gives all of them an
// id without each one declaring the field. Go 1.27 lets a promoted field be a
// key in a composite literal, so the embed does not show up at the call site:
//
//	tgcomp.Button(c, "Save", &tgcomp.ButtonConf{ID: "left_save"})
//
// A Conf must not declare a field named ID of its own. The literal above would
// silently bind to that one and leave Base.ID empty, and nothing would report
// it.
type Base struct {
	// ID identifies the component across runs: it is the key its state is
	// stored under and the id it carries in the DOM. Leave it empty and the
	// component derives one; set it when two components would otherwise
	// derive the same id.
	ID string
}

// base is unexported so that Conf is satisfied only by embedding Base, and so
// that a conf cannot claim to be one by accident.
func (b *Base) base() *Base { return b }

// Conf is what every component conf satisfies, through the Base it embeds. It
// exists so the framework can read a conf's id generically, rather than every
// conf implementing the same getter.
type Conf interface {
	base() *Base
}

// OneConf resolves a component's variadic conf parameter to exactly one conf,
// so a component body can read it without a nil check. An absent conf, and an
// explicit nil, both become a zero conf, which is what "all defaults" means.
//
// More than one is a mistake in the caller rather than a value to resolve:
// there is no sensible way to merge two confs for the same component, so this
// panics instead of picking one.
func OneConf[T any, PC interface {
	*T
	Conf
}](conf []PC) PC {
	if len(conf) > 1 {
		panic("toolgui: a component takes at most one conf")
	}

	if len(conf) == 0 || conf[0] == nil {
		return PC(new(T))
	}

	return conf[0]
}

// SetConfID gives comp the id conf carries, and leaves comp's own id alone
// when the conf carries none. The id is read through the Base every conf
// embeds, so no component writes a getter for it.
func SetConfID(comp Component, conf Conf) {
	id := conf.base().ID
	if id == "" {
		return
	}

	if s, ok := comp.(interface{ SetID(string) }); ok {
		s.SetID(id)
	}
}
