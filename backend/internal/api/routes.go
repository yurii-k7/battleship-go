package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"battleship-go/internal/auth"
	"battleship-go/internal/game"
	"battleship-go/internal/models"
	"battleship-go/internal/websocket"

	"github.com/gin-gonic/gin"
)

type API struct {
	authService *auth.AuthService
	gameService *game.GameService
	hub         *websocket.Hub
	db          *sql.DB
}

func SetupRoutes(router *gin.Engine, db *sql.DB, hub *websocket.Hub) {
	authService := auth.NewAuthService(db, "your-jwt-secret-key")
	gameService := game.NewGameService(db)

	api := &API{
		authService: authService,
		gameService: gameService,
		hub:         hub,
		db:          db,
	}

	// Public routes
	router.POST("/api/auth/register", api.register)
	router.POST("/api/auth/login", api.login)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(api.authMiddleware())
	{
		// User routes
		protected.GET("/user/profile", api.getUserProfile)
		protected.GET("/user/stats", api.getUserStats)

		// Game routes
		protected.POST("/games", api.createGame)
		protected.POST("/games/:id/join", api.joinGame)
		protected.GET("/games", api.getGames)
		protected.GET("/games/:id", api.getGame)
		protected.POST("/games/:id/ships", api.placeShips)
		protected.POST("/games/:id/moves", api.makeMove)
		protected.GET("/games/:id/moves", api.getGameMoves)

		// Chat routes
		protected.POST("/games/:id/chat", api.sendChatMessage)
		protected.GET("/games/:id/chat", api.getChatMessages)

		// Leaderboard
		protected.GET("/leaderboard", api.getLeaderboard)
	}
}

func (a *API) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := a.authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func (a *API) register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := a.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := a.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (a *API) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := a.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (a *API) getUserProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	user, err := a.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (a *API) getUserStats(c *gin.Context) {
	userID := c.GetInt("userID")
	var score models.Score
	err := a.db.QueryRow(`
		SELECT id, player_id, wins, losses, hits, misses, points 
		FROM scores WHERE player_id = $1`, userID).Scan(
		&score.ID, &score.PlayerID, &score.Wins, &score.Losses, &score.Hits, &score.Misses, &score.Points)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stats not found"})
		return
	}
	c.JSON(http.StatusOK, score)
}

func (a *API) createGame(c *gin.Context) {
	userID := c.GetInt("userID")
	game, err := a.gameService.CreateGame(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, game)
}

func (a *API) joinGame(c *gin.Context) {
	userID := c.GetInt("userID")
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	game, err := a.gameService.JoinGame(gameID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}

func (a *API) getGames(c *gin.Context) {
	userID := c.GetInt("userID")
	rows, err := a.db.Query(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at
		FROM games WHERE player1_id = $1 OR player2_id = $1
		ORDER BY updated_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Initialize with empty slice to ensure JSON returns [] instead of null
	games := make([]models.Game, 0)
	for rows.Next() {
		var game models.Game
		err := rows.Scan(&game.ID, &game.Player1ID, &game.Player2ID, &game.Status,
			&game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
		if err != nil {
			continue
		}
		games = append(games, game)
	}

	c.JSON(http.StatusOK, games)
}

func (a *API) getGame(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var game models.Game
	err = a.db.QueryRow(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at 
		FROM games WHERE id = $1`, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status,
		&game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(http.StatusOK, game)
}

func (a *API) placeShips(c *gin.Context) {
	userID := c.GetInt("userID")
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var ships []models.Ship
	if err := c.ShouldBindJSON(&ships); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = a.gameService.PlaceShips(gameID, userID, ships)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ships placed successfully"})
}

func (a *API) makeMove(c *gin.Context) {
	userID := c.GetInt("userID")
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var req struct {
		X int `json:"x" binding:"required,min=0,max=9"`
		Y int `json:"y" binding:"required,min=0,max=9"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	move, err := a.gameService.MakeMove(gameID, userID, req.X, req.Y)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, move)
}

func (a *API) getGameMoves(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	rows, err := a.db.Query(`
		SELECT id, game_id, player_id, x, y, is_hit, ship_id, created_at 
		FROM moves WHERE game_id = $1 ORDER BY created_at`, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Initialize with empty slice to ensure JSON returns [] instead of null
	moves := make([]models.Move, 0)
	for rows.Next() {
		var move models.Move
		err := rows.Scan(&move.ID, &move.GameID, &move.PlayerID, &move.X, &move.Y,
			&move.IsHit, &move.ShipID, &move.CreatedAt)
		if err != nil {
			continue
		}
		moves = append(moves, move)
	}

	c.JSON(http.StatusOK, moves)
}

func (a *API) sendChatMessage(c *gin.Context) {
	userID := c.GetInt("userID")
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var chatMessage models.ChatMessage
	err = a.db.QueryRow(`
		INSERT INTO chat_messages (game_id, player_id, message) 
		VALUES ($1, $2, $3) 
		RETURNING id, game_id, player_id, message, created_at`,
		gameID, userID, req.Message).Scan(
		&chatMessage.ID, &chatMessage.GameID, &chatMessage.PlayerID,
		&chatMessage.Message, &chatMessage.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chatMessage)
}

func (a *API) getChatMessages(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	rows, err := a.db.Query(`
		SELECT id, game_id, player_id, message, created_at 
		FROM chat_messages WHERE game_id = $1 ORDER BY created_at`, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Initialize with empty slice to ensure JSON returns [] instead of null
	messages := make([]models.ChatMessage, 0)
	for rows.Next() {
		var message models.ChatMessage
		err := rows.Scan(&message.ID, &message.GameID, &message.PlayerID,
			&message.Message, &message.CreatedAt)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}

	c.JSON(http.StatusOK, messages)
}

func (a *API) getLeaderboard(c *gin.Context) {
	rows, err := a.db.Query(`
		SELECT s.id, s.player_id, u.username, s.wins, s.losses, s.hits, s.misses, s.points 
		FROM scores s 
		JOIN users u ON s.player_id = u.id 
		ORDER BY s.points DESC LIMIT 10`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Initialize with empty slice to ensure JSON returns [] instead of null
	leaderboard := make([]struct {
		models.Score
		Username string `json:"username"`
	}, 0)

	for rows.Next() {
		var entry struct {
			models.Score
			Username string `json:"username"`
		}
		err := rows.Scan(&entry.ID, &entry.PlayerID, &entry.Username, &entry.Wins,
			&entry.Losses, &entry.Hits, &entry.Misses, &entry.Points)
		if err != nil {
			continue
		}
		leaderboard = append(leaderboard, entry)
	}

	c.JSON(http.StatusOK, leaderboard)
}
