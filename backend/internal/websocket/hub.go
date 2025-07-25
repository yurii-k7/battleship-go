package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
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
				// Remove from game room first
				if client.gameID > 0 && h.gameRooms[client.gameID] != nil {
					delete(h.gameRooms[client.gameID], client)
					if len(h.gameRooms[client.gameID]) == 0 {
						delete(h.gameRooms, client.gameID)
					}
				}
				// Close the send channel to signal writePump to exit
				close(client.send)
				log.Printf("Client unregistered: UserID %d, GameID %d", client.userID, client.gameID)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full or closed, unregister the client
					log.Printf("Failed to send to client UserID %d, unregistering", client.userID)
					h.unregister <- client
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
				// Client's send channel is full or closed, unregister the client
				log.Printf("Failed to send to client UserID %d, unregistering", client.userID)
				h.unregister <- client
			}
		}
	}
}

func (h *Hub) BroadcastToAll(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// Client's send channel is full or closed, unregister the client
			log.Printf("Failed to send to client UserID %d, unregistering", client.userID)
			h.unregister <- client
		}
	}
}

func (h *Hub) SendToUser(userID int, message []byte) {
	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- message:
			default:
				// Client's send channel is full or closed, unregister the client
				log.Printf("Failed to send to client UserID %d, unregistering", client.userID)
				h.unregister <- client
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
	userID := extractUserIDFromToken(r)
	gameIDStr := r.URL.Query().Get("gameId")
	gameID := 0
	if gameIDStr != "" {
		if id, err := strconv.Atoi(gameIDStr); err == nil {
			gameID = id
		}
	}

	log.Printf("WebSocket connection: UserID %d, GameID %d", userID, gameID)

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
		log.Printf("ReadPump closing for UserID %d, GameID %d", c.userID, c.gameID)
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for UserID %d: %v", c.userID, err)
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
	defer func() {
		c.conn.Close()
		log.Printf("WritePump closed for UserID %d, GameID %d", c.userID, c.gameID)
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Channel was closed, send close message and return
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error for UserID %d: %v", c.userID, err)
				return
			}
		}
	}
}

// extractUserIDFromToken extracts the user ID from JWT token in query params
func extractUserIDFromToken(r *http.Request) int {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		return 0
	}

	// Parse the token without verification for now (in production, you should verify)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Printf("Failed to parse JWT token: %v", err)
		return 0
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["user_id"].(float64); ok {
			return int(userID)
		}
	}

	return 0
}
