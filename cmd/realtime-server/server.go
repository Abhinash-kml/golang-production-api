package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Payload string `json:"payload"`
}

func (m Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func main() {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	http.HandleFunc("/realtime", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade connection to websocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal("Error upgrading http to websocket")
		}

		// Extract userid from header and subcribe to it in redis pub sub
		subscribeString := fmt.Sprintf("user:%s", r.Header.Get("userid"))
		sub := rdb.Subscribe(context.Background(), subscribeString)
		iface, err := sub.Receive(context.Background())
		if err != nil {
			fmt.Println("Failed to subcribe to own channel")
		}

		switch iface.(type) {
		case *redis.Subscription:
			fmt.Println("Subscribed")
		case *redis.Message:
			fmt.Println("First message")
		case *redis.Pong:
			fmt.Println("Pong")
		}

		// Read incoming messages and write them to client's connection
		incoming := sub.Channel()
		go func() {
			for message := range incoming {
				stringReader := strings.NewReader(message.Payload)
				var custom Message
				json.NewDecoder(stringReader).Decode(&custom)
				conn.WriteJSON(custom)
			}
		}()

		// Read message from client and route it to redis pub sub
		for {
			networkMessage := Message{}
			err := conn.ReadJSON(&networkMessage)
			if err != nil {
				log.Fatal("Error reading message")
				break
			}

			formattedString := fmt.Sprintf("user:%s", networkMessage.To)
			num, err := rdb.Publish(context.Background(), formattedString, networkMessage).Result()
			if num != 1 || err != nil {
				fmt.Println("Error publishing to user channel. Error:", err.Error())
			}
		}
	})

	args := os.Args[1:]
	port := args[0]
	formattedstring := fmt.Sprintf(":%s", port)
	http.ListenAndServe(formattedstring, nil)
}
