package tgframe

import (
	"errors"
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
		{"Ctrl+plus", "Ctrl+plus"},
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

			if got != tt.want {
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
		{"past the last function key", "f25"},
		{"a modifier twice", "Shift+Shift+o"},
		{"two characters", "CmdOrCtrl+ok"},
		{"the joiner itself", "CmdOrCtrl++"},
		{"not ascii", "CmdOrCtrl+ø"},

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
