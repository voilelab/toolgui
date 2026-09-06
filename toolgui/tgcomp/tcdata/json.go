package tcdata

import (
	"encoding/json"

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
		},
		Value: s,
	}
}

// JSON create a JSON viewer for v.
// If v is a string, it will be treated as a JSON string.
// If v is not a string, it will be serialized to a JSON string.
func JSON(c *tgframe.Container, v any) {
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

	comp := newJSONComponent(serialized)
	c.AddComponent(comp)
}
