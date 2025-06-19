package auth

import (
	"database/sql"
	"errors"
	"time"

	"battleship-go/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *sql.DB
	jwtSecret []byte
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (a *AuthService) Register(username, email, password string) (*models.User, error) {
	// Check if user already exists
	var exists bool
	err := a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)", username, email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert user
	var user models.User
	err = a.db.QueryRow(`
		INSERT INTO users (username, email, password_hash) 
		VALUES ($1, $2, $3) 
		RETURNING id, username, email, created_at, updated_at`,
		username, email, string(hashedPassword)).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Initialize score record
	_, err = a.db.Exec("INSERT INTO scores (player_id) VALUES ($1)", user.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthService) Login(username, password string) (*models.User, string, error) {
	var user models.User
	var hashedPassword string

	err := a.db.QueryRow(`
		SELECT id, username, email, password_hash, created_at, updated_at 
		FROM users WHERE username = $1`, username).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", errors.New("invalid credentials")
		}
		return nil, "", err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := a.GenerateToken(&user)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (a *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (a *AuthService) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	err := a.db.QueryRow(`
		SELECT id, username, email, created_at, updated_at 
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
