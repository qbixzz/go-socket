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
    defer func() {
        delete(clients, ws)
        ws.Close()
    }()

    for {
        var msg Message
        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Printf("error: %v", err)
            break
        }
        log.Printf("Received message: %s", msg.Text)
        if msg.Text == "close" {
            ticker := time.NewTicker(1 * time.Second)
            defer ticker.Stop()
            countdown := 5
            for countdown > 0 {
                <-ticker.C
                countdown--
                countdownMsg := Message{Text: fmt.Sprintf("Closing in %d seconds", countdown), Event: "server-message"}
                broadcast <- countdownMsg
                // err := ws.WriteJSON(countdownMsg)
                // if err != nil {
                //     log.Printf("error: %v", err)
                //     break
                // }
            }
            log.Printf("Closing connection to client: %v", clientId)
            break
        }
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

    r.GET("/ws/", func(c *gin.Context) {
        w := c.Writer
        r := c.Request
        clientId := r.URL.Query().Get("clientId")
        if clientId == "" {
            http.Error(w, "Client ID is required", http.StatusBadRequest)
            return
        }
        handleConnections(c)
    })
    r.POST("/send", handleSendMessage)
    r.GET("/sse", handleSSE)

    go handleMessages()

    r.Run(":8081")
}