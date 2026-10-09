package tgframe

import "encoding/json/jsontext"

const (
	NotifyTypeCreate = 1
	NotifyTypeUpdate = 2
	NotifyTypeDelete = 3

	// NotifyTypeKeep places the node the client already holds at the key, as
	// a create of the same props would. A [Session] sends it in place of a
	// create whose component bytes did not change since the client got them.
	NotifyTypeKeep = 4
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

// notifyPackCreateRaw is a create pack with its component already marshaled,
// which the session needed to compare it anyway.
type notifyPackCreateRaw struct {
	*notifyPackBase
	ParentKey string         `json:"parent_key"`
	Index     int            `json:"index"`
	Key       string         `json:"key"`
	Component jsontext.Value `json:"component"`
}

var _ NotifyPack = &notifyPackKeep{}

type notifyPackKeep struct {
	*notifyPackBase
	ParentKey string `json:"parent_key"`
	Index     int    `json:"index"`
	Key       string `json:"key"`
}

// NewNotifyPackKeep creates a new notify pack for keeping the component the
// client holds at key, placed at index of parentKey.
func NewNotifyPackKeep(parentKey string, index int, key string) *notifyPackKeep {
	return &notifyPackKeep{
		notifyPackBase: &notifyPackBase{
			Type: NotifyTypeKeep,
		},
		ParentKey: parentKey,
		Index:     index,
		Key:       key,
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
	return NewNotifyPackDeleteKey(keyOf(comp))
}

// NewNotifyPackDeleteKey creates a new notify pack for deleting whatever sits
// at key, component and subtree alike. Deleting is idempotent: a key the
// client doesn't have is already in the state the pack asks for, which is what
// lets a container clear a place the previous run wrote and this one has not.
func NewNotifyPackDeleteKey(key string) *notifyPackDelete {
	return &notifyPackDelete{
		notifyPackBase: &notifyPackBase{
			Type: NotifyTypeDelete,
		},
		Key: key,
	}
}
