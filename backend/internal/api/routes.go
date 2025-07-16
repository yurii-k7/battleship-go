package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"battleship-go/internal/auth"
	"battleship-go/internal/cleanup"
	"battleship-go/internal/game"
	"battleship-go/internal/models"
	"battleship-go/internal/websocket"

	"github.com/gin-gonic/gin"
)

type API struct {
	authService    *auth.AuthService
	gameService    *game.GameService
	cleanupService *cleanup.CleanupService
	hub            *websocket.Hub
	db             *sql.DB
}

func SetupRoutes(router *gin.Engine, db *sql.DB, hub *websocket.Hub) {
	authService := auth.NewAuthService(db, "your-jwt-secret-key")
	gameService := game.NewGameService(db)
	cleanupService := cleanup.NewCleanupService(db)

	api := &API{
		authService:    authService,
		gameService:    gameService,
		cleanupService: cleanupService,
		hub:            hub,
		db:             db,
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
		protected.GET("/games/available", api.getAvailableGames)
		protected.GET("/games/:id", api.getGame)
		protected.GET("/games/:id/ships", api.getShips)
		protected.GET("/games/:id/ships/sunk", api.getSunkShips)
		protected.GET("/games/:id/ready", api.checkGameReady)
		protected.POST("/games/:id/ships", api.placeShips)
		protected.POST("/games/:id/moves", api.makeMove)
		protected.GET("/games/:id/moves", api.getGameMoves)

		// Chat routes
		protected.POST("/games/:id/chat", api.sendChatMessage)
		protected.GET("/games/:id/chat", api.getChatMessages)

		// Leaderboard
		protected.GET("/leaderboard", api.getLeaderboard)

		// Admin/Cleanup routes
		protected.GET("/admin/cleanup/status", api.getCleanupStatus)
		protected.POST("/admin/cleanup/run", api.runCleanup)
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

	// Broadcast new game creation to all connected clients
	gameCreateMsg := map[string]interface{}{
		"type":    "new_game_created",
		"data":    game,
		"message": "new_game_available",
	}
	if msgBytes, err := json.Marshal(gameCreateMsg); err == nil {
		a.hub.BroadcastToAll(msgBytes)
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

	// Broadcast game update to all clients in the game
	gameUpdateMsg := map[string]interface{}{
		"type":    "game_update",
		"game_id": gameID,
		"data":    game,
		"message": "player_joined",
	}
	if msgBytes, err := json.Marshal(gameUpdateMsg); err == nil {
		a.hub.BroadcastToGame(gameID, msgBytes)
	}

	c.JSON(http.StatusOK, game)
}

func (a *API) getGames(c *gin.Context) {
	userID := c.GetInt("userID")
	rows, err := a.db.Query(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at
		FROM games WHERE (player1_id = $1 OR player2_id = $1)
		   OR (status = 'waiting' AND player2_id IS NULL AND player1_id != $1)
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

func (a *API) getAvailableGames(c *gin.Context) {
	userID := c.GetInt("userID")
	rows, err := a.db.Query(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at
		FROM games WHERE status = 'waiting' AND player2_id IS NULL AND player1_id != $1
		ORDER BY created_at DESC`, userID)
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

func (a *API) getShips(c *gin.Context) {
	userID := c.GetInt("userID")
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	rows, err := a.db.Query(`
		SELECT id, game_id, player_id, type, size, start_x, start_y, end_x, end_y, is_vertical, is_sunk 
		FROM ships WHERE game_id = $1 AND player_id = $2`, gameID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	ships := make([]models.Ship, 0)
	for rows.Next() {
		var ship models.Ship
		err := rows.Scan(&ship.ID, &ship.GameID, &ship.PlayerID, &ship.Type, &ship.Size,
			&ship.StartX, &ship.StartY, &ship.EndX, &ship.EndY, &ship.IsVertical, &ship.IsSunk)
		if err != nil {
			continue
		}
		ships = append(ships, ship)
	}

	c.JSON(http.StatusOK, ships)
}

func (a *API) getSunkShips(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	rows, err := a.db.Query(`
		SELECT id, game_id, player_id, type, size, start_x, start_y, end_x, end_y, is_vertical, is_sunk 
		FROM ships WHERE game_id = $1 AND is_sunk = true`, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	ships := make([]models.Ship, 0)
	for rows.Next() {
		var ship models.Ship
		err := rows.Scan(&ship.ID, &ship.GameID, &ship.PlayerID, &ship.Type, &ship.Size,
			&ship.StartX, &ship.StartY, &ship.EndX, &ship.EndY, &ship.IsVertical, &ship.IsSunk)
		if err != nil {
			continue
		}
		ships = append(ships, ship)
	}

	c.JSON(http.StatusOK, ships)
}

func (a *API) checkGameReady(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	// Get game info
	var game models.Game
	err = a.db.QueryRow(`
		SELECT id, player1_id, player2_id, status, current_turn, winner_id, created_at, updated_at 
		FROM games WHERE id = $1`, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.Status, &game.CurrentTurn, &game.WinnerID, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// Check if both players have placed ships
	var player1Ships, player2Ships int
	a.db.QueryRow("SELECT COUNT(*) FROM ships WHERE game_id = $1 AND player_id = $2", gameID, game.Player1ID).Scan(&player1Ships)
	if game.Player2ID != nil {
		a.db.QueryRow("SELECT COUNT(*) FROM ships WHERE game_id = $1 AND player_id = $2", gameID, *game.Player2ID).Scan(&player2Ships)
	}

	ready := game.Player2ID != nil && player1Ships == 5 && player2Ships == 5

	c.JSON(http.StatusOK, gin.H{
		"ready":         ready,
		"player1_ships": player1Ships,
		"player2_ships": player2Ships,
		"game_status":   game.Status,
	})
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
		fmt.Printf("Failed to bind JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Placing ships for user %d in game %d: %+v\n", userID, gameID, ships)

	err = a.gameService.PlaceShips(gameID, userID, ships)
	if err != nil {
		fmt.Printf("Failed to place ships: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast ship placement update to all clients in the game
	shipPlacementMsg := map[string]interface{}{
		"type":    "ship_placement_update",
		"game_id": gameID,
		"user_id": userID,
		"message": "ships_placed",
	}
	if msgBytes, err := json.Marshal(shipPlacementMsg); err == nil {
		a.hub.BroadcastToGame(gameID, msgBytes)
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
		X *int `json:"x" binding:"required,min=0,max=9"`
		Y *int `json:"y" binding:"required,min=0,max=9"`
	}


	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Failed to bind move JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Making move: user %d, game %d, position (%d, %d)\n", userID, gameID, *req.X, *req.Y)

	move, err := a.gameService.MakeMove(gameID, userID, *req.X, *req.Y)
	if err != nil {
		fmt.Printf("Failed to make move: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Move successful: %+v\n", move)
	
	// Broadcast game update to all clients in the game
	gameUpdateMsg := map[string]interface{}{
		"type":    "game_update",
		"game_id": gameID,
		"data":    move,
	}
	if msgBytes, err := json.Marshal(gameUpdateMsg); err == nil {
		a.hub.BroadcastToGame(gameID, msgBytes)
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

	// Broadcast chat message to all clients in the game
	chatMsg := map[string]interface{}{
		"type":    "chat",
		"game_id": gameID,
		"data":    chatMessage,
	}
	if msgBytes, err := json.Marshal(chatMsg); err == nil {
		a.hub.BroadcastToGame(gameID, msgBytes)
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

func (a *API) getCleanupStatus(c *gin.Context) {
	count, err := a.cleanupService.GetInactiveGamesCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inactive_games_count": count,
		"cleanup_threshold":    "1 hour",
	})
}

func (a *API) runCleanup(c *gin.Context) {
	err := a.cleanupService.CleanupInactiveGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cleanup completed successfully"})
}
