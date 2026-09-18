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
//	Ctrl+plus
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

// accelModifiers is every modifier, in the order a normalized accelerator
// writes them. Keyed by the lowercased spelling the app may write.
var accelModifiers = map[string]int{
	"cmdorctrl":   0,
	"ctrl":        1,
	"optionoralt": 2,
	"shift":       3,
}

// accelModifierOrder is accelModifiers the other way round: index to spelling.
var accelModifierOrder = []string{
	accelCmdOrCtrl, accelCtrl, accelOptionOrAlt, accelShift,
}

// accelNamedKeys are the keys that are not one character. They are a subset of
// what the desktop menu accepts, cut down to the ones a browser also names:
// the function keys stop at F24 because no `KeyboardEvent.key` goes past it,
// and lock keys are left out because the OS takes them before a page sees
// them.
//
// `plus` is the way to write the `+` that joins the parts.
var accelNamedKeys = func() map[string]bool {
	keys := map[string]bool{
		"backspace": true, "tab": true, "enter": true, "escape": true,
		"left": true, "right": true, "up": true, "down": true,
		"space": true, "delete": true, "home": true, "end": true,
		"page up": true, "page down": true, "plus": true,
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

// ErrAccelerator is what an accelerator the app cannot serve is reported
// with: an unknown modifier or key, a modifier written twice, or nothing to
// press. It is reported through [ErrMenuItem] as well, so a caller checking
// for a bad menu catches it without naming this one.
var ErrAccelerator = tgutil.NewError("invalid accelerator")

// parseAccelerator checks accel and returns its normalized spelling:
// modifiers in a fixed order, each in its canonical case, and the key
// lowercased. Normalizing is what lets two items be compared for declaring
// the same combination in different words.
func parseAccelerator(accel string) (string, error) {
	parts := strings.Split(accel, "+")

	// Everything but the last part is a modifier, and the last one is the
	// key -- which is why `+` itself has to be written `plus`.
	seen := map[int]bool{}
	for _, part := range parts[:len(parts)-1] {
		at, ok := accelModifiers[strings.ToLower(strings.TrimSpace(part))]
		if !ok {
			return "", tgutil.Errorf("%w: `%s` is not a modifier", ErrAccelerator,
				part)
		}

		if seen[at] {
			return "", tgutil.Errorf("%w: `%s` twice", ErrAccelerator, part)
		}

		seen[at] = true
	}

	// CmdOrCtrl already is Control off macOS, so the two together would be
	// one key on every platform but one -- and unpressable on that one.
	if seen[accelModifiers["cmdorctrl"]] && seen[accelModifiers["ctrl"]] {
		return "", tgutil.Errorf(
			"%w: CmdOrCtrl and Ctrl are the same key off macOS", ErrAccelerator)
	}

	key, err := parseAcceleratorKey(parts[len(parts)-1])
	if err != nil {
		return "", tgutil.Errorf("%w", err)
	}

	var out []string
	for at, name := range accelModifierOrder {
		if seen[at] {
			out = append(out, name)
		}
	}

	return strings.Join(append(out, key), "+"), nil
}

// parseAcceleratorKey checks the last part of an accelerator and returns it
// lowercased.
func parseAcceleratorKey(key string) (string, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if accelNamedKeys[key] {
		return key, nil
	}

	// Anything else is one printable ASCII character. A wider set would have
	// to survive a keyboard layout the app does not know about, and a
	// browser reports the character the layout produced rather than the key
	// that was pressed.
	if len(key) == 1 && key[0] > ' ' && key[0] < 0x7f {
		return key, nil
	}

	if key == "" {
		return "", tgutil.Errorf("%w: no key", ErrAccelerator)
	}

	return "", tgutil.Errorf("%w: `%s` is not a key", ErrAccelerator, key)
}
