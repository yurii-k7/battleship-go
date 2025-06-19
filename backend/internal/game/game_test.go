package game

import (
	"database/sql"
	"testing"

	"battleship-go/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			player1_id INTEGER NOT NULL,
			player2_id INTEGER,
			status TEXT DEFAULT 'waiting',
			current_turn INTEGER,
			winner_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE ships (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			player_id INTEGER NOT NULL,
			type TEXT NOT NULL,
			size INTEGER NOT NULL,
			start_x INTEGER NOT NULL,
			start_y INTEGER NOT NULL,
			end_x INTEGER NOT NULL,
			end_y INTEGER NOT NULL,
			is_vertical BOOLEAN NOT NULL,
			is_sunk BOOLEAN DEFAULT FALSE
		);
		
		CREATE TABLE moves (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			player_id INTEGER NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			is_hit BOOLEAN NOT NULL,
			ship_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	// Insert test users
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash) VALUES 
		(1, 'player1', 'player1@test.com', 'hash1'),
		(2, 'player2', 'player2@test.com', 'hash2')
	`)
	require.NoError(t, err)

	return db
}

func TestGameService_CreateGame(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	gameService := NewGameService(db)

	t.Run("successful game creation", func(t *testing.T) {
		game, err := gameService.CreateGame(1)
		
		assert.NoError(t, err)
		assert.NotNil(t, game)
		assert.Equal(t, 1, game.Player1ID)
		assert.Nil(t, game.Player2ID)
		assert.Equal(t, models.GameStatusWaiting, game.Status)
	})
}

func TestGameService_JoinGame(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	gameService := NewGameService(db)

	// Create a game first
	game, err := gameService.CreateGame(1)
	require.NoError(t, err)

	t.Run("successful join", func(t *testing.T) {
		joinedGame, err := gameService.JoinGame(game.ID, 2)
		
		assert.NoError(t, err)
		assert.NotNil(t, joinedGame)
		assert.Equal(t, 1, joinedGame.Player1ID)
		assert.Equal(t, 2, *joinedGame.Player2ID)
		assert.Equal(t, models.GameStatusActive, joinedGame.Status)
		assert.Equal(t, 1, *joinedGame.CurrentTurn) // Player 1 starts
	})

	t.Run("cannot join own game", func(t *testing.T) {
		newGame, err := gameService.CreateGame(1)
		require.NoError(t, err)

		_, err = gameService.JoinGame(newGame.ID, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot join your own game")
	})
}

func TestGameService_ValidateShipPlacement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	gameService := NewGameService(db)

	t.Run("valid ship placement", func(t *testing.T) {
		ships := []models.Ship{
			{Type: "carrier", Size: 5, StartX: 0, StartY: 0, EndX: 4, EndY: 0, IsVertical: false},
			{Type: "battleship", Size: 4, StartX: 0, StartY: 1, EndX: 3, EndY: 1, IsVertical: false},
			{Type: "cruiser", Size: 3, StartX: 0, StartY: 2, EndX: 2, EndY: 2, IsVertical: false},
			{Type: "submarine", Size: 3, StartX: 0, StartY: 3, EndX: 2, EndY: 3, IsVertical: false},
			{Type: "destroyer", Size: 2, StartX: 0, StartY: 4, EndX: 1, EndY: 4, IsVertical: false},
		}

		err := gameService.validateShipPlacement(ships)
		assert.NoError(t, err)
	})

	t.Run("wrong number of ships", func(t *testing.T) {
		ships := []models.Ship{
			{Type: "carrier", Size: 5, StartX: 0, StartY: 0, EndX: 4, EndY: 0, IsVertical: false},
		}

		err := gameService.validateShipPlacement(ships)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must place exactly 5 ships")
	})

	t.Run("overlapping ships", func(t *testing.T) {
		ships := []models.Ship{
			{Type: "carrier", Size: 5, StartX: 0, StartY: 0, EndX: 4, EndY: 0, IsVertical: false},
			{Type: "battleship", Size: 4, StartX: 0, StartY: 0, EndX: 3, EndY: 0, IsVertical: false}, // Overlaps with carrier
			{Type: "cruiser", Size: 3, StartX: 0, StartY: 2, EndX: 2, EndY: 2, IsVertical: false},
			{Type: "submarine", Size: 3, StartX: 0, StartY: 3, EndX: 2, EndY: 3, IsVertical: false},
			{Type: "destroyer", Size: 2, StartX: 0, StartY: 4, EndX: 1, EndY: 4, IsVertical: false},
		}

		err := gameService.validateShipPlacement(ships)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ships cannot overlap")
	})

	t.Run("ship out of bounds", func(t *testing.T) {
		ships := []models.Ship{
			{Type: "carrier", Size: 5, StartX: 6, StartY: 0, EndX: 10, EndY: 0, IsVertical: false}, // Goes beyond board
			{Type: "battleship", Size: 4, StartX: 0, StartY: 1, EndX: 3, EndY: 1, IsVertical: false},
			{Type: "cruiser", Size: 3, StartX: 0, StartY: 2, EndX: 2, EndY: 2, IsVertical: false},
			{Type: "submarine", Size: 3, StartX: 0, StartY: 3, EndX: 2, EndY: 3, IsVertical: false},
			{Type: "destroyer", Size: 2, StartX: 0, StartY: 4, EndX: 1, EndY: 4, IsVertical: false},
		}

		err := gameService.validateShipPlacement(ships)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ship position out of bounds")
	})
}
