package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sbowman/dotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client
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
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
func loginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "auth.html")
}
func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
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
func auth(user string, password string) bool {

	coll := client.Database("chat").Collection("auth")

	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{
		{"username", user},
	}).Decode(&result)

	if err == mongo.ErrNoDocuments {

		fmt.Printf("User not found: %s\n", user)
		return false
	}

	if err != nil {
		panic(err)
	}

	dbPassword, ok := result["password"].(string)
	if !ok {
		fmt.Println("Password field missing or invalid")
		return false
	}

	if checkPassword(dbPassword, password) != true {
		fmt.Println("Wrong password")
		return false
	}

	fmt.Println("Login success:", user)
	return true
}
func register(user string, password string) bool {

	coll := client.Database("chat").Collection("auth")
	hashedpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
	}

	user1 := LoginRequest{Username: user, Password: string(hashedpass)}
	result, err := coll.InsertOne(context.TODO(), user1)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("User not found: %s\n", user)
		return false
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	if err != nil {
		panic(err)
	}

	return true
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)
	ok := auth(req.Username, req.Password)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": ok,
	})
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)
	ok := register(req.Username, req.Password)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": ok,
	})
}
func main() {
	loadHistory()
	err2 := godotenv.Load(".env")
	if err2 != nil {
		log.Fatalf("Error loading .env file: %s", err2)
	}
	fmt.Println(os.Getenv("MONGODB_URI"))
	uri := dotenv.GetString("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	var err error
	client, err = mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "auth.html")
	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	go handleMessages()

	fmt.Println("Server starts on :8080")
	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		panic("error starting server: " + err.Error())
	}
}
