package tgframe

const (
	NotifyTypeCreate = 1
	NotifyTypeUpdate = 2
	NotifyTypeDelete = 3
)

// NotifyPack is the interface that notify packs must implement.
type NotifyPack interface {
	GetType() int
}

type notifyPackBase struct {
	Type int `json:"type"`
}

func (b *notifyPackBase) GetType() int {
	return b.Type
}

// SendNotifyPackFunc is the function type that sends a notify pack to the GUI client.
type SendNotifyPackFunc func(NotifyPack)

var _ NotifyPack = &notifyPackCreate{}

type notifyPackCreate struct {
	*notifyPackBase

	// ParentKey and Index place the component in the node tree; Key is the
	// two joined, which is the component's own key.
	ParentKey string    `json:"parent_key"`
	Index     int       `json:"index"`
	Key       string    `json:"key"`
	Component Component `json:"component"`
}

// NewNotifyPackCreate creates a new notify pack for creating a component.
func NewNotifyPackCreate(parentKey string, index int, key string, comp Component) *notifyPackCreate {
	return &notifyPackCreate{
		notifyPackBase: &notifyPackBase{
			Type: NotifyTypeCreate,
		},
		ParentKey: parentKey,
		Index:     index,
		Key:       key,
		Component: comp,
	}
}

var _ NotifyPack = &notifyPackUpdate{}

type notifyPackUpdate struct {
	*notifyPackBase
	Key       string    `json:"key"`
	Component Component `json:"component"`
}

// NewNotifyPackUpdate creates a new notify pack for updating a component.
func NewNotifyPackUpdate(comp Component) *notifyPackUpdate {
	return &notifyPackUpdate{
		notifyPackBase: &notifyPackBase{
			Type: NotifyTypeUpdate,
		},
		Key:       keyOf(comp),
		Component: comp,
	}
}

var _ NotifyPack = &notifyPackDelete{}

type notifyPackDelete struct {
	*notifyPackBase
	Key string `json:"key"`
}

// NewNotifyPackDelete creates a new notify pack for deleting a component.
func NewNotifyPackDelete(comp Component) *notifyPackDelete {
	return &notifyPackDelete{
		notifyPackBase: &notifyPackBase{
			Type: NotifyTypeDelete,
		},
		Key: keyOf(comp),
	}
}
