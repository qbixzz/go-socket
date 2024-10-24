package main

import (
    "github.com/gin-gonic/gin"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

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
var sseClients = make(map[chan string]bool)

type Message struct {
    Text      string `json:"text"`
    Event     string `json:"event"`
    Room      string `json:"room,omitempty"`
    Recipient string `json:"recipient,omitempty"`
}

func handleConnections(c *gin.Context) {
    w := c.Writer
    r := c.Request
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    clients[ws] = true

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    timer := time.NewTimer(10 * time.Second)
    defer timer.Stop()

    count := 0

    go func() {
        for {
            select {
            case <-ticker.C:
                count++
                msg := Message{Text: fmt.Sprintf("hello %d", count), Event: "server-message"}
                err := ws.WriteJSON(msg)
                if err != nil {
                    log.Printf("error: %v", err)
                    ws.Close()
                    delete(clients, ws)
                    return
                }
            case <-timer.C:
                log.Printf("Disconnecting client after 10 seconds: %v", ws.RemoteAddr())
                ws.Close()
                delete(clients, ws)
                return
            }
        }
    }()

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

    for sseClient := range sseClients {
        sseClient <- msg.Text
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

func handleSendMessage(c *gin.Context) {
    w := c.Writer
    r := c.Request
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

func handleSSE(c *gin.Context) {
    w := c.Writer
    r := c.Request
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    messageChan := make(chan string)
    sseClients[messageChan] = true

    defer func() {
        delete(sseClients, messageChan)
        close(messageChan)
    }()

    for {
        select {
        case msg := <-messageChan:
            fmt.Fprintf(w, "data: %s\n\n", msg)
            flusher.Flush()
        case <-r.Context().Done():
            return
        }
    }
}

func main() {
    r := gin.Default()

    r.GET("/ws", handleConnections)
    r.POST("/send", handleSendMessage)
    r.GET("/sse", handleSSE)

    go handleMessages()

    r.Run(":8080")
}