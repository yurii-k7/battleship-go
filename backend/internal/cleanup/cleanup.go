package cleanup

import (
	"database/sql"
	"log"
	"time"

	"battleship-go/internal/models"
)

type CleanupService struct {
	db *sql.DB
}

func NewCleanupService(db *sql.DB) *CleanupService {
	return &CleanupService{db: db}
}

// StartCleanupScheduler starts a background goroutine that periodically cleans up inactive games
func (c *CleanupService) StartCleanupScheduler() {
	ticker := time.NewTicker(10 * time.Minute) // Run every 10 minutes
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := c.CleanupInactiveGames(); err != nil {
					log.Printf("Error during cleanup: %v", err)
				}
			}
		}
	}()
	log.Println("Cleanup scheduler started - will check for inactive games every 10 minutes")
}

// CleanupInactiveGames removes games that have been inactive for more than 1 hour
func (c *CleanupService) CleanupInactiveGames() error {
	// Define what constitutes an inactive game:
	// 1. Games in 'waiting' status older than 1 hour
	// 2. Games in 'active' status with no moves in the last 1 hour

	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// First, find games to be cleaned up
	rows, err := c.db.Query(`
		SELECT g.id, g.status, g.updated_at, 
		       COALESCE(MAX(m.created_at), g.created_at) as last_activity
		FROM games g
		LEFT JOIN moves m ON g.id = m.game_id
		WHERE g.status != $1 
		GROUP BY g.id, g.status, g.updated_at
		HAVING COALESCE(MAX(m.created_at), g.created_at) < $2
	`, models.GameStatusFinished, oneHourAgo)

	if err != nil {
		return err
	}
	defer rows.Close()

	var gamesToCleanup []int
	for rows.Next() {
		var gameID int
		var status string
		var updatedAt, lastActivity time.Time

		if err := rows.Scan(&gameID, &status, &updatedAt, &lastActivity); err != nil {
			log.Printf("Error scanning game for cleanup: %v", err)
			continue
		}

		log.Printf("Found inactive game %d (status: %s, last activity: %s)",
			gameID, status, lastActivity.Format(time.RFC3339))
		gamesToCleanup = append(gamesToCleanup, gameID)
	}

	// Clean up each game
	for _, gameID := range gamesToCleanup {
		if err := c.cleanupGame(gameID); err != nil {
			log.Printf("Error cleaning up game %d: %v", gameID, err)
		} else {
			log.Printf("Successfully cleaned up inactive game %d", gameID)
		}
	}

	if len(gamesToCleanup) > 0 {
		log.Printf("Cleaned up %d inactive games", len(gamesToCleanup))
	}

	return nil
}

// cleanupGame removes a specific game and all its related data
func (c *CleanupService) cleanupGame(gameID int) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete in the correct order to respect foreign key constraints
	// 1. Delete moves (references ships and games)
	if _, err := tx.Exec("DELETE FROM moves WHERE game_id = $1", gameID); err != nil {
		return err
	}

	// 2. Delete chat messages (references games)
	if _, err := tx.Exec("DELETE FROM chat_messages WHERE game_id = $1", gameID); err != nil {
		return err
	}

	// 3. Delete ships (references games)
	if _, err := tx.Exec("DELETE FROM ships WHERE game_id = $1", gameID); err != nil {
		return err
	}

	// 4. Finally delete the game itself
	if _, err := tx.Exec("DELETE FROM games WHERE id = $1", gameID); err != nil {
		return err
	}

	return tx.Commit()
}

// GetInactiveGamesCount returns the number of games that would be cleaned up
func (c *CleanupService) GetInactiveGamesCount() (int, error) {
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	var count int
	err := c.db.QueryRow(`
		SELECT COUNT(*)
		FROM (
			SELECT g.id
			FROM games g
			LEFT JOIN moves m ON g.id = m.game_id
			WHERE g.status != $1 
			GROUP BY g.id, g.created_at
			HAVING COALESCE(MAX(m.created_at), g.created_at) < $2
		) as inactive_games
	`, models.GameStatusFinished, oneHourAgo).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return count, nil
}
