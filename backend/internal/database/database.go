package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Initialize(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createGamesTable,
		createShipsTable,
		createMovesTable,
		createChatMessagesTable,
		createScoresTable,
		createIndexes,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createGamesTable = `
CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    player1_id INTEGER NOT NULL REFERENCES users(id),
    player2_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'waiting',
    current_turn INTEGER REFERENCES users(id),
    winner_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createShipsTable = `
CREATE TABLE IF NOT EXISTS ships (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id),
    player_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(20) NOT NULL,
    size INTEGER NOT NULL,
    start_x INTEGER NOT NULL,
    start_y INTEGER NOT NULL,
    end_x INTEGER NOT NULL,
    end_y INTEGER NOT NULL,
    is_vertical BOOLEAN NOT NULL,
    is_sunk BOOLEAN DEFAULT FALSE
);`

const createMovesTable = `
CREATE TABLE IF NOT EXISTS moves (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id),
    player_id INTEGER NOT NULL REFERENCES users(id),
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    is_hit BOOLEAN NOT NULL,
    ship_id INTEGER REFERENCES ships(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createChatMessagesTable = `
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id),
    player_id INTEGER NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createScoresTable = `
CREATE TABLE IF NOT EXISTS scores (
    id SERIAL PRIMARY KEY,
    player_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    hits INTEGER DEFAULT 0,
    misses INTEGER DEFAULT 0,
    points INTEGER DEFAULT 0
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_players ON games(player1_id, player2_id);
CREATE INDEX IF NOT EXISTS idx_moves_game ON moves(game_id);
CREATE INDEX IF NOT EXISTS idx_ships_game_player ON ships(game_id, player_id);
CREATE INDEX IF NOT EXISTS idx_chat_game ON chat_messages(game_id);
CREATE INDEX IF NOT EXISTS idx_scores_points ON scores(points DESC);
ALTER TABLE moves DROP CONSTRAINT IF EXISTS moves_game_id_x_y_key;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'moves_game_player_position_key') THEN
        ALTER TABLE moves ADD CONSTRAINT moves_game_player_position_key UNIQUE (game_id, player_id, x, y);
    END IF;
END $$;`
