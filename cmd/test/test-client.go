package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientMessage struct {
	SenderID   string `json:"senderid"`
	ReceiverID string `json:"receiverid"`
	Payload    string `json:"payload"`
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

	wg.Wait()
}

func ReadFromStdIn(conn *websocket.Conn, wg *sync.WaitGroup, senderid string) {
	defer wg.Done()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Split(input, ":") // Format = receiverid: Payload. Example - 111: Hi bye
		receiverid := parts[0]
		payload := parts[1]
		message := ClientMessage{
			SenderID:   senderid,
			ReceiverID: receiverid,
			Payload:    payload,
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
		var message ClientMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s: %s\n", message.SenderID, message.Payload)
	}
}
