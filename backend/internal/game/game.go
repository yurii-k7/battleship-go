package game

import (
	"database/sql"
	"errors"
	"fmt"

	"battleship-go/internal/models"
)

type GameService struct {
	db *sql.DB
}

func NewGameService(db *sql.DB) *GameService {
	return &GameService{db: db}
}

func (g *GameService) CreateGame(playerID int) (*models.Game, error) {
	var game models.Game
	err := g.db.QueryRow(`
		INSERT INTO games (player1_id, status) 
		VALUES ($1, $2) 
		RETURNING id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at`,
		playerID, models.GameStatusWaiting).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status, &game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (g *GameService) JoinGame(gameID, playerID int) (*models.Game, error) {
	// Check if game exists and is waiting for players
	var game models.Game
	err := g.db.QueryRow(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at 
		FROM games WHERE id = $1`, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status, &game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if game.Status != models.GameStatusWaiting {
		return nil, errors.New("game is not available for joining")
	}

	if game.Player1ID == playerID {
		return nil, errors.New("cannot join your own game")
	}

	// Update game with second player
	err = g.db.QueryRow(`
		UPDATE games SET player2_id = $1, status = $2, current_turn = $3, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $4 
		RETURNING id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at`,
		playerID, models.GameStatusActive, game.Player1ID, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status, &game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func (g *GameService) PlaceShips(gameID, playerID int, ships []models.Ship) error {
	// Validate ship placement
	if err := g.validateShipPlacement(ships); err != nil {
		return err
	}

	// Check if ships are already placed for this player
	var count int
	err := g.db.QueryRow("SELECT COUNT(*) FROM ships WHERE game_id = $1 AND player_id = $2", gameID, playerID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("ships already placed")
	}

	// Insert ships
	for _, ship := range ships {
		_, err := g.db.Exec(`
			INSERT INTO ships (game_id, player_id, type, size, start_x, start_y, end_x, end_y, is_vertical) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			gameID, playerID, ship.Type, ship.Size, ship.StartX, ship.StartY, ship.EndX, ship.EndY, ship.IsVertical)
		if err != nil {
			return err
		}
	}

	// Check if both players have now placed their ships
	err = g.checkAndStartGame(gameID)
	if err != nil {
		return err
	}

	return nil
}

func (g *GameService) checkAndStartGame(gameID int) error {
	// Get game info
	var game models.Game
	err := g.db.QueryRow(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at 
		FROM games WHERE id = $1`, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status, &game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		return err
	}

	// Only proceed if game is active and has both players
	if game.Status != models.GameStatusActive || game.Player2ID == nil {
		return nil
	}

	// Check if both players have placed ships
	var player1Ships, player2Ships int
	g.db.QueryRow("SELECT COUNT(*) FROM ships WHERE game_id = $1 AND player_id = $2", gameID, game.Player1ID).Scan(&player1Ships)
	g.db.QueryRow("SELECT COUNT(*) FROM ships WHERE game_id = $1 AND player_id = $2", gameID, *game.Player2ID).Scan(&player2Ships)

	// If both players have placed ships and current_turn is NULL, set it to player1
	if player1Ships == 5 && player2Ships == 5 && game.CurrentTurn == nil {
		_, err = g.db.Exec("UPDATE games SET current_turn = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", game.Player1ID, gameID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GameService) MakeMove(gameID, playerID, x, y int) (*models.Move, error) {
	// Check if it's the player's turn
	var game models.Game
	err := g.db.QueryRow(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id 
		FROM games WHERE id = $1`, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status, &game.CurrentTurn, &game.WinnerID)
	if err != nil {
		return nil, err
	}

	if game.Status != models.GameStatusActive {
		return nil, errors.New("game is not active")
	}

	if game.CurrentTurn == nil || *game.CurrentTurn != playerID {
		return nil, errors.New("not your turn")
	}

	// Check if move already exists by this player
	var existingMoveCount int
	err = g.db.QueryRow("SELECT COUNT(*) FROM moves WHERE game_id = $1 AND player_id = $2 AND x = $3 AND y = $4", gameID, playerID, x, y).Scan(&existingMoveCount)
	if err != nil {
		return nil, err
	}
	if existingMoveCount > 0 {
		return nil, errors.New("position already targeted")
	}

	// Determine opponent
	opponentID := game.Player1ID
	if playerID == game.Player1ID {
		opponentID = *game.Player2ID
	}

	// Check for hit
	var shipID *int
	var isHit bool
	err = g.db.QueryRow(`
		SELECT id FROM ships 
		WHERE game_id = $1 AND player_id = $2 
		AND ((is_vertical = true AND start_x = $3 AND $4 BETWEEN start_y AND end_y) 
		OR (is_vertical = false AND start_y = $4 AND $3 BETWEEN start_x AND end_x))`,
		gameID, opponentID, x, y).Scan(&shipID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	isHit = (shipID != nil)

	// Insert move
	var move models.Move
	err = g.db.QueryRow(`
		INSERT INTO moves (game_id, player_id, x, y, is_hit, ship_id) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, game_id, player_id, x, y, is_hit, ship_id, created_at`,
		gameID, playerID, x, y, isHit, shipID).Scan(
		&move.ID, &move.GameID, &move.PlayerID, &move.X, &move.Y, &move.IsHit, &move.ShipID, &move.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Check if ship is sunk
	if isHit && shipID != nil {
		g.checkAndUpdateSunkShip(gameID, *shipID)
	}

	// Check for game end
	if g.checkGameEnd(gameID, opponentID) {
		g.endGame(gameID, playerID)
	} else {
		// Switch turns
		nextPlayer := opponentID
		if isHit {
			nextPlayer = playerID // Player gets another turn on hit
		}
		g.db.Exec("UPDATE games SET current_turn = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", nextPlayer, gameID)
	}

	return &move, nil
}

func (g *GameService) validateShipPlacement(ships []models.Ship) error {
	if len(ships) != 5 {
		return errors.New("must place exactly 5 ships")
	}

	// Check ship types and sizes
	expectedShips := map[string]int{
		"carrier":    5,
		"battleship": 4,
		"cruiser":    3,
		"submarine":  3,
		"destroyer":  2,
	}

	shipCounts := make(map[string]int)
	for _, ship := range ships {
		shipCounts[ship.Type]++
		if expectedSize, exists := expectedShips[ship.Type]; !exists || ship.Size != expectedSize {
			return fmt.Errorf("invalid ship type or size: %s", ship.Type)
		}
	}

	// Validate ship counts
	for shipType, expectedCount := range map[string]int{"carrier": 1, "battleship": 1, "cruiser": 1, "submarine": 1, "destroyer": 1} {
		if shipCounts[shipType] != expectedCount {
			return fmt.Errorf("incorrect number of %s ships", shipType)
		}
	}

	// Check for overlaps and valid positions
	positions := make(map[string]bool)
	for _, ship := range ships {
		if ship.StartX < 0 || ship.StartX > 9 || ship.StartY < 0 || ship.StartY > 9 ||
			ship.EndX < 0 || ship.EndX > 9 || ship.EndY < 0 || ship.EndY > 9 {
			return errors.New("ship position out of bounds")
		}

		// Generate all positions for this ship
		if ship.IsVertical {
			for y := ship.StartY; y <= ship.EndY; y++ {
				pos := fmt.Sprintf("%d,%d", ship.StartX, y)
				if positions[pos] {
					return errors.New("ships cannot overlap")
				}
				positions[pos] = true
			}
		} else {
			for x := ship.StartX; x <= ship.EndX; x++ {
				pos := fmt.Sprintf("%d,%d", x, ship.StartY)
				if positions[pos] {
					return errors.New("ships cannot overlap")
				}
				positions[pos] = true
			}
		}
	}

	return nil
}

func (g *GameService) checkAndUpdateSunkShip(gameID, shipID int) {
	// Get ship details
	var ship models.Ship
	err := g.db.QueryRow(`
		SELECT id, game_id, player_id, type, size, start_x, start_y, end_x, end_y, is_vertical, is_sunk
		FROM ships WHERE id = $1`, shipID).Scan(
		&ship.ID, &ship.GameID, &ship.PlayerID, &ship.Type, &ship.Size,
		&ship.StartX, &ship.StartY, &ship.EndX, &ship.EndY, &ship.IsVertical, &ship.IsSunk)
	if err != nil {
		return
	}

	// Count hits on each position of this ship
	hitCount := 0
	if ship.IsVertical {
		for y := ship.StartY; y <= ship.EndY; y++ {
			var moveExists int
			g.db.QueryRow(`
				SELECT COUNT(*) FROM moves 
				WHERE game_id = $1 AND x = $2 AND y = $3 AND is_hit = true AND player_id != $4`,
				gameID, ship.StartX, y, ship.PlayerID).Scan(&moveExists)
			if moveExists > 0 {
				hitCount++
			}
		}
	} else {
		for x := ship.StartX; x <= ship.EndX; x++ {
			var moveExists int
			g.db.QueryRow(`
				SELECT COUNT(*) FROM moves 
				WHERE game_id = $1 AND x = $2 AND y = $3 AND is_hit = true AND player_id != $4`,
				gameID, x, ship.StartY, ship.PlayerID).Scan(&moveExists)
			if moveExists > 0 {
				hitCount++
			}
		}
	}

	fmt.Printf("Ship %d: %d hits out of %d size\n", shipID, hitCount, ship.Size)

	if hitCount >= ship.Size {
		fmt.Printf("Ship %d is sunk!\n", shipID)
		g.db.Exec("UPDATE ships SET is_sunk = true WHERE id = $1", shipID)
	}
}

func (g *GameService) checkGameEnd(gameID, playerID int) bool {
	var sunkShips, totalShips int
	g.db.QueryRow(`
		SELECT COUNT(CASE WHEN is_sunk THEN 1 END), COUNT(*) 
		FROM ships WHERE game_id = $1 AND player_id = $2`, gameID, playerID).Scan(&sunkShips, &totalShips)

	fmt.Printf("Game end check for player %d: %d sunk ships out of %d total\n", playerID, sunkShips, totalShips)
	gameEnded := sunkShips == totalShips
	fmt.Printf("Game ended: %v\n", gameEnded)

	return gameEnded
}

func (g *GameService) endGame(gameID, winnerID int) {
	g.db.Exec(`
		UPDATE games SET status = $1, winner_id = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3`, models.GameStatusFinished, winnerID, gameID)

	// Update scores
	g.updatePlayerScore(winnerID, true)

	// Get loser ID
	var loserID int
	g.db.QueryRow(`
		SELECT CASE WHEN player1_id = $1 THEN player2_id ELSE player1_id END 
		FROM games WHERE id = $2`, winnerID, gameID).Scan(&loserID)
	g.updatePlayerScore(loserID, false)
}

func (g *GameService) updatePlayerScore(playerID int, won bool) {
	if won {
		g.db.Exec(`
			UPDATE scores SET wins = wins + 1, points = points + 100 
			WHERE player_id = $1`, playerID)
	} else {
		g.db.Exec(`
			UPDATE scores SET losses = losses + 1 
			WHERE player_id = $1`, playerID)
	}
}
