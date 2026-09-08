package tcdata

import (
	"encoding/json"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
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

// JSONConf is the configuration for the JSON component.
type JSONConf struct {
	tgframe.Base
}

// JSON create a JSON viewer for v.
// If v is a string, it will be treated as a JSON string.
// If v is not a string, it will be serialized to a JSON string.
func JSON(c *tgframe.Container, v any, conf ...*JSONConf) {
	cf := tgframe.OneConf("JSON", conf)

	serialized, err := serializeJSON(v)
	if err != nil {
		c.Fail(tgutil.Errorf("%w", err))
		return
	}

	comp := newJSONComponent(serialized)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}

func serializeJSON(v any) (string, error) {
	if res, ok := v.(string); ok {
		// check if the string is a valid JSON
		var js map[string]any
		if err := json.Unmarshal([]byte(res), &js); err != nil {
			return "", err
		}

		return res, nil
	}

	bs, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}
