package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sbowman/dotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))
var client *mongo.Client
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var messagesColl *mongo.Collection

type User struct {
	Username string `bson:"username" json:"username"`
}
type Message struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string        `json:"username"`
	Recipient string        `json:"recipient"`
	Message   string        `json:"message"`
	Time      time.Time     `bson:"time" json:"time"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var clients = make(map[string]*websocket.Conn)
var broadcast = make(chan Message)

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}
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
	token := r.URL.Query().Get("token")

	username, err := validateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[username] = conn

	for {

		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, username)
			return
		}
		msg.Username = username
		msg.Time = time.Now()
		_, err = messagesColl.InsertOne(context.TODO(), msg)
		if err != nil {
			log.Printf("Mongo insert error: %v", err)
		}

		if msg.Recipient == "" {
			broadcast <- msg
		} else {
			if recipientConn, ok := clients[msg.Recipient]; ok {
				recipientConn.WriteJSON(msg)
			}

			if senderConn, ok := clients[msg.Username]; ok {
				senderConn.WriteJSON(msg)
			}
		}
	}
}
func getMessages(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	if len(token) > 7 {
		token = token[7:]
	}

	username, err := validateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	filter := bson.M{
		"$or": []bson.M{
			{"recipient": ""},
			{"username": username},
			{"recipient": username},
		},
	}
	cursor, err := messagesColl.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cursor.Close(context.TODO())

	var msgs []Message
	cursor.All(context.TODO(), &msgs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}
func handleMessages() {
	for {
		msg := <-broadcast

		for username, conn := range clients {
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				conn.Close()
				delete(clients, username)
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
	if ok {
		token, _ := generateToken(req.Username)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"token":   token,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
		})

	}
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
func getUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	coll := client.Database("chat").Collection("auth")

	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var dbUsers []User
	if err := cursor.All(context.TODO(), &dbUsers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var users []string
	for _, u := range dbUsers {
		users = append(users, u.Username)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func validateToken(tokenString string) (string, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
	)

	if err != nil || !token.Valid {
		return "", err
	}

	return claims.Username, nil
}
func main() {

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
	messagesColl = client.Database("chat").Collection("messages")
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/users", getUsers)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "auth.html")
	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/messages", getMessages)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	go handleMessages()

	fmt.Println("Server starts on :8080")
	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		panic("error starting server: " + err.Error())
	}
}
