package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in development
	},
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	gameRooms  map[int]map[*Client]bool // gameID -> clients
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int
	gameID int
}

type Message struct {
	Type    string      `json:"type"`
	GameID  int         `json:"game_id,omitempty"`
	UserID  int         `json:"user_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		gameRooms:  make(map[int]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if client.gameID > 0 {
				if h.gameRooms[client.gameID] == nil {
					h.gameRooms[client.gameID] = make(map[*Client]bool)
				}
				h.gameRooms[client.gameID][client] = true
			}
			log.Printf("Client registered: UserID %d, GameID %d", client.userID, client.gameID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if client.gameID > 0 && h.gameRooms[client.gameID] != nil {
					delete(h.gameRooms[client.gameID], client)
					if len(h.gameRooms[client.gameID]) == 0 {
						delete(h.gameRooms, client.gameID)
					}
				}
				log.Printf("Client unregistered: UserID %d, GameID %d", client.userID, client.gameID)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastToGame(gameID int, message []byte) {
	if gameClients, exists := h.gameRooms[gameID]; exists {
		for client := range gameClients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(gameClients, client)
			}
		}
	}
}

func (h *Hub) SendToUser(userID int, message []byte) {
	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Extract user and game info from query parameters
	userID := 0 // This should be extracted from JWT token
	gameID := 0 // This should be extracted from query parameters

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		gameID: gameID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "chat":
			// Broadcast chat message to game room
			if c.gameID > 0 {
				c.hub.BroadcastToGame(c.gameID, messageBytes)
			}
		case "move":
			// Handle game move
			if c.gameID > 0 {
				c.hub.BroadcastToGame(c.gameID, messageBytes)
			}
		case "join_game":
			// Handle joining a game
			if gameID, ok := msg.Data.(float64); ok {
				c.gameID = int(gameID)
				if c.hub.gameRooms[c.gameID] == nil {
					c.hub.gameRooms[c.gameID] = make(map[*Client]bool)
				}
				c.hub.gameRooms[c.gameID][c] = true
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		}
	}
}
