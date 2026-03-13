package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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

func main() {
	args := os.Args[1:]
	uid := args[0]
	port := args[1]

	dialer := websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	uidHeader := http.Header{
		"uid": []string{uid},
	}

	conn, _, err := dialer.Dial(fmt.Sprintf("ws://localhost:%s/realtime", port), uidHeader)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go ReadFromStdIn(conn, &wg, uid)
	wg.Add(1)
	go ReadFromConnection(conn, &wg)
	//go SendPeriodicHeartbeat(conn)

	wg.Wait()
}

func SendPeriodicHeartbeat(conn *websocket.Conn) {
	ticker := time.NewTicker(time.Second * 30)

	for range ticker.C {
		err := conn.WriteMessage(websocket.PongMessage, nil)
		if err != nil {
			log.Fatal("Failed to write pong message to server. Error:", err.Error())
		}
	}
}

func ReadFromStdIn(conn *websocket.Conn, wg *sync.WaitGroup, senderid string) {
	defer wg.Done()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Split(input, ":") // Format = receiverid: Payload. Example - 111: Hi bye
		receiverid := parts[0]
		payload := parts[1]
		payloadbytes, _ := json.Marshal(payload)
		message := Envelope{
			Header: Header{
				SenderID:      senderid,
				RecieverID:    receiverid,
				CorrelationID: "aaaaaa",
				Category:      int(CategoryMessage),
			},
			Type:      TypeMessage,
			Data:      json.RawMessage(payloadbytes),
			Timestamp: time.Now(),
		}

		err := conn.WriteJSON(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func ReadFromConnection(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		var envelope Envelope
		err := conn.ReadJSON(&envelope)
		if err != nil {
			log.Fatal(err)
		}

		var message string
		json.Unmarshal(envelope.Data, &message)
		fmt.Printf("%s - %s: %s\n", envelope.Timestamp.Format(time.DateTime), envelope.Header.SenderID, message)
	}
}
