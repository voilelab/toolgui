package tgwails

import "testing"

func TestConfDefaults(t *testing.T) {
	def := DefaultConf()

	conf := (&Conf{Title: "Mine"}).withDefaults()
	if conf.Title != "Mine" {
		t.Fatalf("expect the given title, got %q", conf.Title)
	}
	if conf.Width != def.Width || conf.Height != def.Height {
		t.Fatalf("expect the default size, got %dx%d", conf.Width, conf.Height)
	}
	// Compare the colours, not the pointers.
	if *conf.Background != *def.Background {
		t.Fatalf("expect the default background, got %v", conf.Background)
	}

	// A window is resizable unless the caller opts out, so the zero value of
	// Conf is a usable config.
	if conf.DisableResize {
		t.Fatal("expect a resizable window by default")
	}

	if (*Conf)(nil).withDefaults().Title != def.Title {
		t.Fatal("expect a nil conf to fall back to the defaults")
	}
}
