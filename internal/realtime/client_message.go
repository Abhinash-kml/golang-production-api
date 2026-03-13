package realtime

import (
	"encoding/json"
	"time"
)

type MessageCategory int

const (
	CategoryMessage MessageCategory = iota + 1
	CategoryNotification
	CategorySystem
)

const (
	TypeMessage = iota + 1
	TypeMessageReply
	TypeMessageFwd
	TypeMessageReact
)

type Header struct {
	SourceID      string `json:"src"`
	SenderID      string `json:"sid"`
	RecieverID    string `json:"rid"`
	CorrelationID string `json:"cid"`
	Category      int    `json:"cat"`
}

type Envelope struct {
	Header    Header          `json:"header"`
	Type      int             `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"ts"`
}

func (e *Envelope) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Envelope) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

type ChatMessage struct {
	Body string `json:"body"`
}

type MessageReply struct {
	Body     string `json:"body"`
	ParentID string `json:"ref_id"`
}

type MessageForward struct {
	OriginID string `json:"org_id"`
	ParentID string `json:"ref_id"`
	Body     string `json:"body"`
}

type Notification struct {
	Title string `json:"title"`
	Icon  int    `json:"icon"`
	Level int    `json:"level"`
	Body  int    `json:"body"`
}
