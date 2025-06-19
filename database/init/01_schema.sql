-- Create database schema for Battleship game

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    player1_id INTEGER NOT NULL REFERENCES users(id),
    player2_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'waiting' CHECK (status IN ('waiting', 'active', 'finished')),
    current_turn INTEGER REFERENCES users(id),
    winner_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ships table
CREATE TABLE IF NOT EXISTS ships (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('carrier', 'battleship', 'cruiser', 'submarine', 'destroyer')),
    size INTEGER NOT NULL CHECK (size BETWEEN 2 AND 5),
    start_x INTEGER NOT NULL CHECK (start_x BETWEEN 0 AND 9),
    start_y INTEGER NOT NULL CHECK (start_y BETWEEN 0 AND 9),
    end_x INTEGER NOT NULL CHECK (end_x BETWEEN 0 AND 9),
    end_y INTEGER NOT NULL CHECK (end_y BETWEEN 0 AND 9),
    is_vertical BOOLEAN NOT NULL,
    is_sunk BOOLEAN DEFAULT FALSE
);

-- Moves table
CREATE TABLE IF NOT EXISTS moves (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id INTEGER NOT NULL REFERENCES users(id),
    x INTEGER NOT NULL CHECK (x BETWEEN 0 AND 9),
    y INTEGER NOT NULL CHECK (y BETWEEN 0 AND 9),
    is_hit BOOLEAN NOT NULL,
    ship_id INTEGER REFERENCES ships(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, x, y)
);

-- Chat messages table
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id INTEGER NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Scores table
CREATE TABLE IF NOT EXISTS scores (
    id SERIAL PRIMARY KEY,
    player_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
    wins INTEGER DEFAULT 0 CHECK (wins >= 0),
    losses INTEGER DEFAULT 0 CHECK (losses >= 0),
    hits INTEGER DEFAULT 0 CHECK (hits >= 0),
    misses INTEGER DEFAULT 0 CHECK (misses >= 0),
    points INTEGER DEFAULT 0 CHECK (points >= 0)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_players ON games(player1_id, player2_id);
CREATE INDEX IF NOT EXISTS idx_games_created_at ON games(created_at);
CREATE INDEX IF NOT EXISTS idx_moves_game ON moves(game_id);
CREATE INDEX IF NOT EXISTS idx_moves_position ON moves(game_id, x, y);
CREATE INDEX IF NOT EXISTS idx_ships_game_player ON ships(game_id, player_id);
CREATE INDEX IF NOT EXISTS idx_chat_game ON chat_messages(game_id);
CREATE INDEX IF NOT EXISTS idx_chat_created_at ON chat_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_scores_points ON scores(points DESC);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_games_updated_at BEFORE UPDATE ON games
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
