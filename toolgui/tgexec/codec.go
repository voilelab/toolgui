package tgexec

import (
	"github.com/voilelab/toolgui/toolgui/tgjson"

	"golang.org/x/net/websocket"
)

// jsonCodec is what every pack on an update connection goes over.
// websocket.JSON cannot be given codec options and encodes with the v1
// semantics, so using it would leave the server decoding events one way and
// encoding packs another.
var jsonCodec = websocket.Codec{
	Marshal: func(v any) ([]byte, byte, error) {
		bs, err := tgjson.Marshal(v)
		return bs, websocket.TextFrame, err
	},
	Unmarshal: func(data []byte, _ byte, v any) error {
		return tgjson.Unmarshal(data, v)
	},
}
