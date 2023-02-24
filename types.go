package tinypubsub

import "encoding/json"

type ActionType = string
type Topic = string
type ID = string
type DataType = string

const (
	ActionPublish     ActionType = "publish"
	ActionSubscribe   ActionType = "subscribe"
	ActionUnsubscribe ActionType = "unsubscribe"
)

func IsValidAction(a ActionType) bool {
	return a == ActionPublish || a == ActionSubscribe || a == ActionUnsubscribe
}

const (
	DataError DataType = "error"
	DataMsg   DataType = "message"
)

const (
	PeerIDParamKey   = "peerid"
	WebSocketUrlPath = "/socket"
)

type ActionMessage struct {
	Action ActionType `json:"action"`
	Topic  Topic      `json:"topic"`
	Data   []byte     `json:"data"`
}

func (m *ActionMessage) FromBytes(b []byte) error {
	return json.Unmarshal(b, m)
}

func (m ActionMessage) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

type DataMessage struct {
	Typ       DataType `json:"type"`
	Topic     Topic    `json:"topic"`
	Publisher ID       `json:"publisher"`
	Data      []byte   `json:"data"`
}

func (m *DataMessage) FromBytes(b []byte) error {
	return json.Unmarshal(b, m)
}

func (m DataMessage) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}
