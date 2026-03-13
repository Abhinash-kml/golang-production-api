package realtime

import (
	"encoding/json"
	"time"
)

type MessageCategory int
type MessageType int
type ReceiptStatus int

const (
	CategoryMessage MessageCategory = iota + 1
	CategoryBroadcast
	CategoryNotification
	CategorySystem
)

const (
	TypeMessage = iota + 1
	TypeMessageReply
	TypeMessageFwd
	TypeMessageReact
)

const (
	StatusSent ReceiptStatus = iota + 1
	StatusDelivered
	StatusRead
)

type Header struct {
	SourceID      string          `json:"src"`
	SenderID      string          `json:"sid"` // Set by client
	RecieverID    string          `json:"rid"` // Set by client
	CorrelationID string          `json:"cid"` // Set by client
	Category      MessageCategory `json:"cat"` // For server side routing & processing
	Hops          int             `json:"hops"`
}

type Envelope struct {
	Header    Header          `json:"header"`
	Type      MessageType     `json:"type"` // For client side processing
	Data      json.RawMessage `json:"data"` // Set by client
	Timestamp time.Time       `json:"ts"`   // Set by client
}

func NewEnvelope(sourceid, senderid, receiverid, correlationid string, category MessageCategory, messagetype MessageType, data json.RawMessage, timestamp time.Time) Envelope {
	return Envelope{
		Header: Header{
			SourceID:      sourceid,
			SenderID:      senderid,
			RecieverID:    receiverid,
			CorrelationID: correlationid,
			Category:      category,
			Hops:          1,
		},
		Type:      messagetype,
		Data:      data,
		Timestamp: timestamp,
	}
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

type ReadReceipt struct {
	CorrelationID string        `json:"cid"`
	Status        ReceiptStatus `json:"status"`
}
