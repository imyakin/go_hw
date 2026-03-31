package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/render"
	"github.com/imyakin/go_hw/internal/repository"
)

type CreateMoveRequest struct {
	GameID  int    `json:"game_id" binding:"required"`
	FromRow int    `json:"from_row"`
	FromCol int    `json:"from_col"`
	ToRow   int    `json:"to_row"`
	ToCol   int    `json:"to_col"`
	// Alternative: chess notation like "e2-e4"
	Notation string `json:"notation"`
}

type UpdateMoveRequest struct {
	GameID  int    `json:"game_id" binding:"required"`
	FromRow int    `json:"from_row" binding:"required"`
	FromCol int    `json:"from_col" binding:"required"`
	ToRow   int    `json:"to_row" binding:"required"`
	ToCol   int    `json:"to_col" binding:"required"`
	Piece   string `json:"piece" binding:"required"`
}

type MoveResponse struct {
	ID       int    `json:"id"`
	GameID   int    `json:"game_id"`
	FromRow  int    `json:"from_row"`
	FromCol  int    `json:"from_col"`
	ToRow    int    `json:"to_row"`
	ToCol    int    `json:"to_col"`
	Piece    string `json:"piece"`
	Notation string `json:"notation"`
	Player   string `json:"player,omitempty"`
}

func moveToResponse(m *model.Move) MoveResponse {
	playerName := ""
	if m.Player != nil {
		playerName = m.Player.GetDisplayName()
	}
	return MoveResponse{
		ID:       m.ID,
		GameID:   m.GameID,
		FromRow:  m.From.Row,
		FromCol:  m.From.Col,
		ToRow:    m.To.Row,
		ToCol:    m.To.Col,
		Piece:    m.Piece,
		Notation: m.GetNotation(),
		Player:   playerName,
	}
}

// CreateMove creates a new move in a game
// @Summary Create a new move
// @Description Makes a move in an active chess game. Use either notation (e.g. "e2-e4") or row/col coordinates.
// @Tags moves
// @Accept json
// @Produce json
// @Param input body CreateMoveRequest true "Move parameters"
// @Success 201 {object} MoveResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /api/moves [post]
func CreateMove(c *gin.Context) {
	var req CreateMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game := repository.GetGameByID(req.GameID)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	game.Mu.Lock()

	if !game.IsInProgress() {
		game.Mu.Unlock()
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "game is not in progress"})
		return
	}

	var move *model.Move

	if req.Notation != "" {
		var err error
		move, err = render.ParseMove(req.Notation, game.CurrentPlayer, game.Board)
		if err != nil {
			game.Mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		piece := game.Board.GetCell(req.FromRow, req.FromCol)
		if piece == "" {
			game.Mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": "нет фигуры на начальной позиции"})
			return
		}
		move = model.NewMove(req.FromRow, req.FromCol, req.ToRow, req.ToCol, game.CurrentPlayer, piece)
	}

	if !move.IsValid(game.Board) {
		game.Mu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid move coordinates"})
		return
	}

	if err := render.ApplyMove(game.Board, move); err != nil {
		game.Mu.Unlock()
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	game.MakeMove(move)
	move.GameID = req.GameID
	gameID := game.ID
	game.Mu.Unlock()

	repository.Store(move)

	// Save updated game state to CSV (outside game.Mu to avoid deadlock with saveGamesCSV)
	repository.UpdateGame(gameID, func(g *model.Game) {})

	c.JSON(http.StatusCreated, moveToResponse(move))
}

// ListMoves returns all moves, optionally filtered by game_id
// @Summary List all moves
// @Description Returns a list of all moves. Can be filtered by game_id query parameter.
// @Tags moves
// @Produce json
// @Param game_id query int false "Filter by game ID"
// @Success 200 {array} MoveResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/moves [get]
func ListMoves(c *gin.Context) {
	allMoves := repository.GetMoves()

	// Optional filter by game_id
	gameIDStr := c.Query("game_id")
	if gameIDStr != "" {
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
			return
		}
		filtered := make([]MoveResponse, 0)
		for _, m := range allMoves {
			if m.GameID == gameID {
				filtered = append(filtered, moveToResponse(m))
			}
		}
		c.JSON(http.StatusOK, filtered)
		return
	}

	result := make([]MoveResponse, 0, len(allMoves))
	for _, m := range allMoves {
		result = append(result, moveToResponse(m))
	}
	c.JSON(http.StatusOK, result)
}

// GetMove returns a move by ID
// @Summary Get a move by ID
// @Description Returns a single move by its ID
// @Tags moves
// @Produce json
// @Param id path int true "Move ID"
// @Success 200 {object} MoveResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/moves/{id} [get]
func GetMove(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	move := repository.GetMoveByID(id)
	if move == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "move not found"})
		return
	}

	c.JSON(http.StatusOK, moveToResponse(move))
}

// UpdateMove updates an existing move
// @Summary Update a move
// @Description Updates an existing move by ID
// @Tags moves
// @Accept json
// @Produce json
// @Param id path int true "Move ID"
// @Param input body UpdateMoveRequest true "Move update parameters"
// @Success 200 {object} MoveResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/moves/{id} [put]
func UpdateMove(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	found := repository.UpdateMove(id, func(m *model.Move) {
		m.GameID = req.GameID
		m.From = model.Position{Row: req.FromRow, Col: req.FromCol}
		m.To = model.Position{Row: req.ToRow, Col: req.ToCol}
		m.Piece = req.Piece
	})

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("move %d not found", id)})
		return
	}

	move := repository.GetMoveByID(id)
	c.JSON(http.StatusOK, moveToResponse(move))
}
