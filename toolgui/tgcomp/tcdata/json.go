package tcdata

import (
	"encoding/json"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &jsonComponent{}
var jsonComponentName = "json_component"

type jsonComponent struct {
	*tgframe.BaseComponent
	Value string `json:"value"`
}

func newJSONComponent(s string) *jsonComponent {
	return &jsonComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: jsonComponentName,
			// The viewer keeps which nodes are collapsed, so the component
			// needs a name of its own to keep that across runs.
			ID: tcutil.HashedID(jsonComponentName, []byte(s)),
		},
		Value: s,
	}
}

// JSON create a JSON viewer for v.
// If v is a string, it will be treated as a JSON string.
// If v is not a string, it will be serialized to a JSON string.
func JSON(c *tgframe.Container, v any) {
	c.AddComponent(newJSONComponent(serializeJSON(v)))
}

// JSONWithID create a JSON viewer with a user specific id.
func JSONWithID(c *tgframe.Container, v any, id string) {
	comp := newJSONComponent(serializeJSON(v))
	comp.SetID(id)
	c.AddComponent(comp)
}

func serializeJSON(v any) string {
	var serialized string

	if res, ok := v.(string); ok {
		// check if the string is a valid JSON
		var js map[string]any
		err := json.Unmarshal([]byte(res), &js)
		if err != nil {
			panic(err)
		}

		serialized = res
	} else {
		bs, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		serialized = string(bs)
	}

	return serialized
}
