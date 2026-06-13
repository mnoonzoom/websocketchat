package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var messages []Message

func loadHistory() {
	file, err := os.ReadFile("msghistory.json")
	if err != nil {
		return
	}
	json.Unmarshal(file, &messages)
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {

		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}

		msg.Time = time.Now().Format("15:04:05")
		messages = append(messages, msg)

		file, err := os.Create("msghistory.json")
		if err != nil {
			log.Println(err)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(messages); err != nil {
			log.Println(err)
		}
		broadcast <- msg
	}
}
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
func main() {
	loadHistory()
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/msghistory.json", http.FileServer(http.Dir(".")))
	go handleMessages()

	fmt.Println("Server starts on :8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic("error starting server: " + err.Error())
	}
}
