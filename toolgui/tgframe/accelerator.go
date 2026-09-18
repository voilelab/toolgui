package tgframe

import (
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// An accelerator is the key combination that fires a menu item without the
// menu being opened. It is declared once, in a platform independent spelling,
// and each carrier serves it its own way: on the desktop the OS dispatches it
// off the native menu item, in the browser the shell listens for the keystroke
// itself.
//
// The spelling is modifiers and then one key, joined by `+`:
//
//	CmdOrCtrl+O
//	CmdOrCtrl+Shift+F5
//	Ctrl+Shift+/
//
// Modifiers and keys are case insensitive; [MenuNode.Accelerator] holds the
// normalized form.

// Accelerator modifiers. CmdOrCtrl and OptionOrAlt are what an app should
// reach for: they are the modifier the platform's own menus use, so one
// declaration reads right on macOS and off it. Ctrl and Shift are the same
// key everywhere.
const (
	accelCmdOrCtrl   = "CmdOrCtrl"
	accelOptionOrAlt = "OptionOrAlt"
	accelShift       = "Shift"
	accelCtrl        = "Ctrl"
)

// accelNamedKeys are the keys that are not one character. They are a subset of
// what the desktop menu accepts, cut down to the ones a browser also names:
// the function keys stop at F24 because no `KeyboardEvent.key` goes past it,
// and lock keys are left out because the OS takes them before a page sees
// them.
var accelNamedKeys = func() map[string]bool {
	keys := map[string]bool{
		"backspace": true, "tab": true, "enter": true, "escape": true,
		"left": true, "right": true, "up": true, "down": true,
		"space": true, "delete": true, "home": true, "end": true,
		"page up": true, "page down": true,
	}

	for _, f := range []string{
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11",
		"f12", "f13", "f14", "f15", "f16", "f17", "f18", "f19", "f20", "f21",
		"f22", "f23", "f24",
	} {
		keys[f] = true
	}

	return keys
}()

// accelPunctuation is the punctuation an accelerator may name: the character a
// key is printed with when Shift is not held. Letters and digits are the rest
// of the set.
const accelPunctuation = "`-=[]\\;',./"

// accelShifted maps a character that needs Shift to the key it shares. An
// accelerator names a key rather than the character the key produces, so `?`
// is refused and `Shift+/` is the way to say it -- which is also the only
// spelling the two carriers agree on, since a browser reports `?` for that
// keystroke and the desktop is handed `/`.
//
// `+` is in here as well, so the character joining the parts is never a key.
var accelShifted = map[byte]byte{
	'~': '`', '!': '1', '@': '2', '#': '3', '$': '4', '%': '5',
	'^': '6', '&': '7', '*': '8', '(': '9', ')': '0', '_': '-',
	'+': '=', '{': '[', '}': ']', '|': '\\', ':': ';', '"': '\'',
	'<': ',', '>': '.', '?': '/',
}

// ErrAccelerator is what an accelerator the app cannot serve is reported
// with: an unknown modifier or key, a modifier written twice, or nothing to
// press. It is reported through [ErrMenuItem] as well, so a caller checking
// for a bad menu catches it without naming this one.
var ErrAccelerator = tgutil.NewError("invalid accelerator")

// accelerator is a parsed declaration.
type accelerator struct {
	cmdOrCtrl, ctrl, optionOrAlt, shift bool

	// key is lowercased: one of accelNamedKeys, a letter, a digit, or one of
	// accelPunctuation.
	key string
}

// String is the normalized spelling: the modifiers in one order whatever
// order they were written in, so two items declaring the same combination in
// different words compare equal.
func (a accelerator) String() string {
	var parts []string
	for _, m := range []struct {
		on   bool
		name string
	}{
		{a.cmdOrCtrl, accelCmdOrCtrl},
		{a.ctrl, accelCtrl},
		{a.optionOrAlt, accelOptionOrAlt},
		{a.shift, accelShift},
	} {
		if m.on {
			parts = append(parts, m.name)
		}
	}

	return strings.Join(append(parts, a.key), "+")
}

// accelWhere names the platform two declarations meet on, in the message
// about them meeting there.
const accelWhere = "off macOS"

// keystroke is the keys actually held down, which is what makes two
// declarations one shortcut rather than two. `CmdOrCtrl+o` and `Ctrl+o` read
// differently and are Control and `o` both, so the second would never fire.
//
// It is the reading off macOS, because that is the only place two
// declarations can meet: macOS is where CmdOrCtrl and Ctrl are different
// keys, so two accelerators alike there are alike as text too, and the
// message about them says so instead.
func (a accelerator) keystroke() string {
	out := ""
	for _, m := range []struct {
		on bool
		c  byte
	}{
		{a.ctrl || a.cmdOrCtrl, 'C'}, {a.optionOrAlt, 'A'}, {a.shift, 'S'},
	} {
		if m.on {
			out += string(m.c)
		}
	}

	return out + " " + a.key
}

// parseAccelerator checks accel and returns it parsed.
func parseAccelerator(accel string) (accelerator, error) {
	var out accelerator

	parts := strings.Split(accel, "+")

	// Everything but the last part is a modifier, and the last one is the
	// key -- which is why `+` is not a key an accelerator can name.
	for _, part := range parts[:len(parts)-1] {
		var at *bool

		switch strings.ToLower(strings.TrimSpace(part)) {
		case "cmdorctrl":
			at = &out.cmdOrCtrl
		case "ctrl":
			at = &out.ctrl
		case "optionoralt":
			at = &out.optionOrAlt
		case "shift":
			at = &out.shift
		default:
			return out, tgutil.Errorf("%w: `%s` is not a modifier",
				ErrAccelerator, part)
		}

		if *at {
			return out, tgutil.Errorf("%w: `%s` twice", ErrAccelerator, part)
		}

		*at = true
	}

	// CmdOrCtrl already is Control off macOS, so the two together would be
	// one key on every platform but one -- and unpressable on that one.
	if out.cmdOrCtrl && out.ctrl {
		return out, tgutil.Errorf(
			"%w: CmdOrCtrl and Ctrl are the same key off macOS", ErrAccelerator)
	}

	key, err := parseAcceleratorKey(parts[len(parts)-1])
	if err != nil {
		return out, tgutil.Errorf("%w", err)
	}

	out.key = key

	return out, nil
}

// parseAcceleratorKey checks the last part of an accelerator and returns it
// lowercased.
func parseAcceleratorKey(key string) (string, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if accelNamedKeys[key] {
		return key, nil
	}

	if key == "" {
		return "", tgutil.Errorf("%w: no key", ErrAccelerator)
	}

	// Anything else is one key of the standard layout, named by the
	// character it is printed with. A wider set would have to survive a
	// keyboard layout the app does not know about, and a browser reports the
	// character the layout produced rather than the key that was pressed.
	if len(key) == 1 {
		c := key[0]
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' ||
			strings.IndexByte(accelPunctuation, c) >= 0 {

			return key, nil
		}

		// A shifted character names no key of its own, so say which key it
		// is on rather than turning it away as unknown.
		if unshifted, ok := accelShifted[c]; ok {
			return "", tgutil.Errorf(
				"%w: `%s` is Shift and the `%c` key, so write `Shift+%c`",
				ErrAccelerator, key, unshifted, unshifted)
		}
	}

	return "", tgutil.Errorf("%w: `%s` is not a key", ErrAccelerator, key)
}
