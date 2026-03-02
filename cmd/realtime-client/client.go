package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Payload string `json:"payload"`
}

func main() {
	args := os.Args[1:]
	from := args[0]
	to := args[1]
	port := args[2]
	connectionString := fmt.Sprintf("ws://localhost:%s/realtime", port)

	dialer := websocket.Dialer{}
	userid := http.Header{
		"userid": []string{from},
	}
	conn, _, err := dialer.Dial(connectionString, userid)
	if err != nil {
		log.Fatal("Error:", err.Error())
	}

	go ReadFromConnection(conn)
	ReadFromStdIn(conn, from, to)
}

func ReadFromStdIn(conn *websocket.Conn, from, to string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		message := Message{
			From:    from,
			To:      to,
			Payload: input,
		}

		err := conn.WriteJSON(message)
		if err != nil {
			log.Fatal("Error writing to websocket connection. Error:", err.Error())
		}
	}
}

func ReadFromConnection(conn *websocket.Conn) {
	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Fatal("Error reading incoming message from connection")
		}
		fmt.Printf("%s: %s", message.From, message.Payload)
	}
}
