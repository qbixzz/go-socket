package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]bool)
var rooms = make(map[string]map[*websocket.Conn]bool)
var broadcast = make(chan Message)

type Message struct {
    Text      string `json:"text"`
    Event     string `json:"event"`
    Room      string `json:"room,omitempty"`
    Recipient string `json:"recipient,omitempty"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    clientId := r.URL.Query().Get("clientId")
    if clientId == "" {
        http.Error(w, "Client ID is required", http.StatusBadRequest)
        return
    }
    log.Println("New client connected", clientId)
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    clients[ws] = true

    for {
        var msg Message
        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Printf("error: %v", err)
            delete(clients, ws)
            break
        }
        log.Printf("Received message: %s", msg.Text)

        if msg.Room != "" {
            if rooms[msg.Room] == nil {
                rooms[msg.Room] = make(map[*websocket.Conn]bool)
            }
            rooms[msg.Room][ws] = true
        }

        broadcast <- msg
    }
}

func handleMessages() {
    for {
        msg := <-broadcast

        if msg.Room != "" {
            handleRoomMessage(msg)
        } else if msg.Recipient != "" {
            handleRecipientMessage(msg)
        } else {
            handleBroadcastMessage(msg)
        }
    }
}

func handleRoomMessage(msg Message) {
    for client := range rooms[msg.Room] {
        if err := sendMessageToClient(client, msg); err != nil {
            handleClientError(client, msg.Room)
        }
    }
}

func handleRecipientMessage(msg Message) {
    for client := range clients {
        if client.RemoteAddr().String() == msg.Recipient {
            if err := sendMessageToClient(client, msg); err != nil {
                handleClientError(client, "")
            }
            break
        }
    }
}

func handleBroadcastMessage(msg Message) {
    for client := range clients {
        if err := sendMessageToClient(client, msg); err != nil {
            handleClientError(client, "")
        }
    }
}

func sendMessageToClient(client *websocket.Conn, msg Message) error {
    return client.WriteJSON(msg)
}

func handleClientError(client *websocket.Conn, room string) {
    client.Close()
    delete(clients, client)
    if room != "" {
        delete(rooms[room], client)
    }
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
    var msg Message
    err := json.NewDecoder(r.Body).Decode(&msg)
    if err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    msg.Event = "server-message"
    log.Printf("Sended message via API: %s", msg.Text)
    broadcast <- msg
    w.WriteHeader(http.StatusOK)
}

func main() {
    fs := http.FileServer(http.Dir("./public"))
    http.Handle("/", fs)
    http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
        clientId := r.URL.Query().Get("clientId")
        if clientId == "" {
            http.Error(w, "Client ID is required", http.StatusBadRequest)
            return
        }
        handleConnections(w, r)
    })
    http.HandleFunc("/send", handleSendMessage)

    go handleMessages()

    fmt.Println("HTTP server started on :8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}