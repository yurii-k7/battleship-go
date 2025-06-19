-- Sample data for development and testing

-- Insert sample users (passwords are hashed for 'password123')
INSERT INTO users (username, email, password_hash) VALUES 
('player1', 'player1@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
('player2', 'player2@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
('admiral', 'admiral@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
('captain', 'captain@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi')
ON CONFLICT (username) DO NOTHING;

-- Initialize scores for sample users
INSERT INTO scores (player_id, wins, losses, hits, misses, points) 
SELECT id, 0, 0, 0, 0, 0 FROM users 
ON CONFLICT (player_id) DO NOTHING;

-- Insert a sample completed game for demonstration
INSERT INTO games (player1_id, player2_id, status, winner_id) 
SELECT 
    (SELECT id FROM users WHERE username = 'player1'),
    (SELECT id FROM users WHERE username = 'player2'),
    'finished',
    (SELECT id FROM users WHERE username = 'player1')
WHERE NOT EXISTS (SELECT 1 FROM games);

-- Update sample scores
UPDATE scores SET 
    wins = 1, 
    hits = 15, 
    misses = 5, 
    points = 150 
WHERE player_id = (SELECT id FROM users WHERE username = 'player1');

UPDATE scores SET 
    losses = 1, 
    hits = 12, 
    misses = 8, 
    points = 120 
WHERE player_id = (SELECT id FROM users WHERE username = 'player2');

UPDATE scores SET 
    wins = 2, 
    losses = 1, 
    hits = 25, 
    misses = 10, 
    points = 250 
WHERE player_id = (SELECT id FROM users WHERE username = 'admiral');

UPDATE scores SET 
    wins = 1, 
    losses = 2, 
    hits = 18, 
    misses = 15, 
    points = 180 
WHERE player_id = (SELECT id FROM users WHERE username = 'captain');
