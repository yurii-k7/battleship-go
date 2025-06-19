package auth

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use in-memory SQLite for testing
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
		
		CREATE TABLE scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			player_id INTEGER UNIQUE NOT NULL,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			hits INTEGER DEFAULT 0,
			misses INTEGER DEFAULT 0,
			points INTEGER DEFAULT 0
		);
	`)
	require.NoError(t, err)

	return db
}

func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authService := NewAuthService(db, "test-secret")

	t.Run("successful registration", func(t *testing.T) {
		user, err := authService.Register("testuser", "test@example.com", "password123")
		
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NotZero(t, user.ID)
	})

	t.Run("duplicate username", func(t *testing.T) {
		_, err := authService.Register("testuser", "test2@example.com", "password123")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := authService.Register("testuser2", "test@example.com", "password123")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
	})
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authService := NewAuthService(db, "test-secret")

	// Register a user first
	_, err := authService.Register("testuser", "test@example.com", "password123")
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		user, token, err := authService.Login("testuser", "password123")
		
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("invalid username", func(t *testing.T) {
		_, _, err := authService.Login("nonexistent", "password123")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("invalid password", func(t *testing.T) {
		_, _, err := authService.Login("testuser", "wrongpassword")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestAuthService_GenerateAndValidateToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authService := NewAuthService(db, "test-secret")

	// Register a user
	user, err := authService.Register("testuser", "test@example.com", "password123")
	require.NoError(t, err)

	t.Run("generate and validate token", func(t *testing.T) {
		token, err := authService.GenerateToken(user)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := authService.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Username, claims.Username)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := authService.ValidateToken("invalid-token")
		assert.Error(t, err)
	})
}
