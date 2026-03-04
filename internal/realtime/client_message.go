package realtime

type ClientMessage struct {
	SenderID   string `json:"senderid"`
	ReceiverID string `json:"receiverid"`
	Payload    string `json:"payload"`
}

func (m *ClientMessage) MarshalBinary() ([]byte, error) {

}

func (m *ClientMessage) UnmarshalBinary(data []byte) error {

}
