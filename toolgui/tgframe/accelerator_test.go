package tgframe

import (
	"errors"
	"strings"
	"testing"
)

// TestParseAccelerator is the normalizing itself: what an app may write, and
// the one spelling it is stored under.
func TestParseAccelerator(t *testing.T) {
	for _, tt := range []struct {
		accel, want string
	}{
		{"CmdOrCtrl+O", "CmdOrCtrl+o"},
		{"cmdorctrl+o", "CmdOrCtrl+o"},
		{"CMDORCTRL+O", "CmdOrCtrl+o"},
		{"o", "o"},
		{"F5", "f5"},
		{"f24", "f24"},
		{"Ctrl+Shift+/", "Ctrl+Shift+/"},
		{"CmdOrCtrl+=", "CmdOrCtrl+="},
		{"CmdOrCtrl+page up", "CmdOrCtrl+page up"},
		{"Shift+escape", "Shift+escape"},
		{"OptionOrAlt+Shift+7", "OptionOrAlt+Shift+7"},

		// The modifiers come back in one order however they were written, so
		// two items declaring the same combination read as the same string.
		{"Shift+CmdOrCtrl+s", "CmdOrCtrl+Shift+s"},
		{"Shift+OptionOrAlt+Ctrl+s", "Ctrl+OptionOrAlt+Shift+s"},
	} {
		t.Run(tt.accel, func(t *testing.T) {
			got, err := parseAccelerator(tt.accel)
			if err != nil {
				t.Fatalf("parseAccelerator(%q) = %v", tt.accel, err)
			}

			if got.String() != tt.want {
				t.Errorf("parseAccelerator(%q) = %q, want %q",
					tt.accel, got, tt.want)
			}
		})
	}
}

func TestParseAcceleratorInvalid(t *testing.T) {
	for _, tt := range []struct{ name, accel string }{
		{"empty", ""},
		{"no key", "CmdOrCtrl+"},
		{"unknown modifier", "Super+o"},
		{"unknown named key", "CmdOrCtrl+numlock"},
		{"a key that was dropped", "Ctrl+plus"},
		{"past the last function key", "f25"},
		{"a modifier twice", "Shift+Shift+o"},
		{"two characters", "CmdOrCtrl+ok"},
		{"the joiner itself", "CmdOrCtrl++"},
		{"not ascii", "CmdOrCtrl+ø"},

		// A shifted character names no key of its own, and the two carriers
		// would not agree on it: a browser reports `?` where the desktop is
		// handed `/`.
		{"a shifted character", "CmdOrCtrl+?"},
		{"a shifted digit", "CmdOrCtrl+!"},

		// Off macOS these are one key, so the combination is unpressable
		// on every platform but the one it reads differently on.
		{"CmdOrCtrl with Ctrl", "CmdOrCtrl+Ctrl+o"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAccelerator(tt.accel)
			if err == nil {
				t.Fatalf("parseAccelerator(%q) = %q, want an error",
					tt.accel, got)
			}

			if !errors.Is(err, ErrAccelerator) {
				t.Errorf("parseAccelerator(%q) = %v, want ErrAccelerator",
					tt.accel, err)
			}
		})
	}
}

// TestAcceleratorShiftedKeyNamesTheKey: a shifted character is refused with
// the spelling to use instead, rather than as an unknown key. The two are on
// one key, and Shift is what the declaration is missing.
func TestAcceleratorShiftedKeyNamesTheKey(t *testing.T) {
	_, err := parseAccelerator("CmdOrCtrl+?")
	if err == nil {
		t.Fatal("parseAccelerator(`CmdOrCtrl+?`) returned no error")
	}

	if !strings.Contains(err.Error(), "Shift+/") {
		t.Errorf("error = %q, want it to name `Shift+/`", err)
	}
}

// TestAcceleratorKeystroke is what the duplicate check compares: the keys
// actually held down, which is not the declaration. CmdOrCtrl is Control off
// macOS, so two declarations that read differently are one keystroke there.
func TestAcceleratorKeystroke(t *testing.T) {
	keystroke := func(accel string) string {
		parsed, err := parseAccelerator(accel)
		if err != nil {
			t.Fatalf("parseAccelerator(%q) = %v", accel, err)
		}

		return parsed.keystroke()
	}

	if keystroke("CmdOrCtrl+o") != keystroke("Ctrl+o") {
		t.Errorf("CmdOrCtrl+o is %q and Ctrl+o is %q, want one keystroke",
			keystroke("CmdOrCtrl+o"), keystroke("Ctrl+o"))
	}

	// Every other modifier is the same key on every platform, so it tells
	// two accelerators apart wherever they run.
	for _, accel := range []string{
		"CmdOrCtrl+Shift+o", "CmdOrCtrl+OptionOrAlt+o", "CmdOrCtrl+p",
	} {
		if keystroke(accel) == keystroke("CmdOrCtrl+o") {
			t.Errorf("%s reads as CmdOrCtrl+o", accel)
		}
	}
}
