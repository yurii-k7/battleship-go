package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Game struct {
	ID          int       `json:"id" db:"id"`
	Player1ID   int       `json:"player1_id" db:"player1_id"`
	Player2ID   *int      `json:"player2_id" db:"player2_id"`
	Status      string    `json:"status" db:"status"` // waiting, active, finished
	CurrentTurn *int      `json:"current_turn" db:"current_turn"`
	WinnerID    *int      `json:"winner_id" db:"winner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Ship struct {
	ID        int    `json:"id" db:"id"`
	GameID    int    `json:"game_id" db:"game_id"`
	PlayerID  int    `json:"player_id" db:"player_id"`
	Type      string `json:"type" db:"type"` // carrier, battleship, cruiser, submarine, destroyer
	Size      int    `json:"size" db:"size"`
	StartX    int    `json:"start_x" db:"start_x"`
	StartY    int    `json:"start_y" db:"start_y"`
	EndX      int    `json:"end_x" db:"end_x"`
	EndY      int    `json:"end_y" db:"end_y"`
	IsVertical bool  `json:"is_vertical" db:"is_vertical"`
	IsSunk    bool   `json:"is_sunk" db:"is_sunk"`
}

type Move struct {
	ID        int       `json:"id" db:"id"`
	GameID    int       `json:"game_id" db:"game_id"`
	PlayerID  int       `json:"player_id" db:"player_id"`
	X         int       `json:"x" db:"x"`
	Y         int       `json:"y" db:"y"`
	IsHit     bool      `json:"is_hit" db:"is_hit"`
	ShipID    *int      `json:"ship_id" db:"ship_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ChatMessage struct {
	ID        int       `json:"id" db:"id"`
	GameID    int       `json:"game_id" db:"game_id"`
	PlayerID  int       `json:"player_id" db:"player_id"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Score struct {
	ID       int `json:"id" db:"id"`
	PlayerID int `json:"player_id" db:"player_id"`
	Wins     int `json:"wins" db:"wins"`
	Losses   int `json:"losses" db:"losses"`
	Hits     int `json:"hits" db:"hits"`
	Misses   int `json:"misses" db:"misses"`
	Points   int `json:"points" db:"points"`
}

// Game status constants
const (
	GameStatusWaiting  = "waiting"
	GameStatusActive   = "active"
	GameStatusFinished = "finished"
)

// Ship types and sizes
var ShipTypes = map[string]int{
	"carrier":    5,
	"battleship": 4,
	"cruiser":    3,
	"submarine":  3,
	"destroyer":  2,
}
