package tgexec

import (
	"encoding/json"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ManifestIcon is one entry of [Manifest].Icons.
type ManifestIcon struct {
	// Src is the icon url, resolved against the app root.
	Src string `json:"src"`

	// Type is the mime type of the icon, e.g. "image/png".
	Type string `json:"type,omitempty"`

	// Sizes is a space separated list of sizes, e.g. "192x192".
	Sizes string `json:"sizes,omitempty"`

	// Purpose is the icon purpose, e.g. "maskable" or "any".
	Purpose string `json:"purpose,omitempty"`
}

// Manifest is the web app manifest [WebExecutor] serves at /manifest.json.
// Empty fields are left out, so only what is set reaches the browser.
type Manifest struct {
	Name        string `json:"name,omitempty"`
	ShortName   string `json:"short_name,omitempty"`
	Description string `json:"description,omitempty"`

	Icons []ManifestIcon `json:"icons,omitempty"`

	StartURL string `json:"start_url,omitempty"`
	Scope    string `json:"scope,omitempty"`

	// Display is how the app is presented, e.g. "standalone" or "browser".
	Display     string `json:"display,omitempty"`
	Orientation string `json:"orientation,omitempty"`

	ThemeColor      string `json:"theme_color,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`

	Lang string `json:"lang,omitempty"`
	Dir  string `json:"dir,omitempty"`

	// Extra holds manifest members this struct doesn't name, such as
	// "shortcuts" or "categories". They are merged into the served json, and
	// a key here wins over the field of the same name.
	Extra map[string]any `json:"-"`
}

// DefaultManifest return the base of the manifest served when the app sets
// none. The app title, when it has one, names that manifest.
func DefaultManifest() *Manifest {
	// No icons: toolgui has no icon to hand out, and an app's own belongs to
	// the app.
	return &Manifest{
		Name:            "ToolGUI App",
		ShortName:       "ToolGUI App",
		StartURL:        ".",
		Display:         "standalone",
		ThemeColor:      "#000000",
		BackgroundColor: "#ffffff",
	}
}

// MarshalJSON merge Extra into the named members.
func (m *Manifest) MarshalJSON() ([]byte, error) {
	// The alias drops MarshalJSON, so this doesn't recurse.
	type manifest Manifest

	bs, err := json.Marshal((*manifest)(m))
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if len(m.Extra) == 0 {
		return bs, nil
	}

	members := map[string]any{}
	err = json.Unmarshal(bs, &members)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	for name, value := range m.Extra {
		members[name] = value
	}

	bs, err = json.Marshal(members)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return bs, nil
}
