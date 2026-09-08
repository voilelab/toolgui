package tcmisc

import (
	"encoding/json"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// barPack is a notify pack cut down to what this test asserts on.
type barPack struct {
	Type      int `json:"type"`
	Component struct {
		Value int `json:"value"`
	} `json:"component"`
}

// importer keeps the bar rather than moving it along where it was drawn. That
// the handle is an exported type is what lets it be a field here and a
// parameter of step; the test fails by not compiling if it stops being one.
type importer struct {
	bar *ProgressBarHandle
}

func step(bar *ProgressBarHandle, done int) {
	bar.SetValue(done)
}

func TestProgressBarHandleIsCarriedAround(t *testing.T) {
	var packs []barPack
	container := tgframe.NewContainer("test", tgframe.NewState(),
		func(pack tgframe.NotifyPack) {
			bs, err := json.Marshal(pack)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var one barPack
			if err := json.Unmarshal(bs, &one); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			packs = append(packs, one)
		})

	im := &importer{bar: ProgressBar(container, 0, "Importing…")}
	step(im.bar, 50)
	im.bar.Remove()

	if len(packs) != 3 {
		t.Fatalf("got %d packs, want 3 (create, update, delete)", len(packs))
	}

	if packs[1].Type != tgframe.NotifyTypeUpdate || packs[1].Component.Value != 50 {
		t.Errorf("second pack = type %d value %d, want update to 50",
			packs[1].Type, packs[1].Component.Value)
	}

	if packs[2].Type != tgframe.NotifyTypeDelete {
		t.Errorf("third pack = type %d, want delete", packs[2].Type)
	}
}
