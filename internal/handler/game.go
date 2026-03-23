package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/render"
	"github.com/imyakin/go_hw/internal/repository"
)

type CreateGameRequest struct {
	WhitePlayerName string `json:"white_player_name" binding:"required"`
	BlackPlayerName string `json:"black_player_name" binding:"required"`
	BoardSize       int    `json:"board_size" binding:"required,min=1"`
}

type UpdateGameRequest struct {
	WhitePlayerName string `json:"white_player_name" binding:"required"`
	BlackPlayerName string `json:"black_player_name" binding:"required"`
	BoardSize       int    `json:"board_size" binding:"required,min=1"`
	Status          string `json:"status" binding:"required"`
}

type GameResponse struct {
	ID                 int        `json:"id"`
	WhitePlayerName    string     `json:"white_player_name"`
	BlackPlayerName    string     `json:"black_player_name"`
	BoardSize          int        `json:"board_size"`
	Status             string     `json:"status"`
	CurrentPlayerColor string     `json:"current_player_color"`
	WinnerColor        string     `json:"winner_color,omitempty"`
	Cells              [][]string `json:"cells"`
}

func gameToResponse(g *model.Game) GameResponse {
	g.Mu.RLock()
	defer g.Mu.RUnlock()

	currentColor := ""
	if g.CurrentPlayer != nil {
		currentColor = string(g.CurrentPlayer.Color)
	}
	winnerColor := ""
	if g.Winner != nil {
		winnerColor = string(g.Winner.Color)
	}

	return GameResponse{
		ID:                 g.ID,
		WhitePlayerName:    g.WhitePlayer.Name,
		BlackPlayerName:    g.BlackPlayer.Name,
		BoardSize:          g.Board.Size,
		Status:             string(g.Status),
		CurrentPlayerColor: currentColor,
		WinnerColor:        winnerColor,
		Cells:              g.Board.Cells,
	}
}

func CreateGame(c *gin.Context) {
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game := model.NewGame(req.WhitePlayerName, req.BlackPlayerName, req.BoardSize)
	render.PlacePieces(game.Board, req.BoardSize)
	game.Start()

	repository.Store(game)
	repository.Store(game.Board)
	repository.Store(game.WhitePlayer)
	repository.Store(game.BlackPlayer)

	c.JSON(http.StatusCreated, gameToResponse(game))
}

func ListGames(c *gin.Context) {
	games := repository.GetGames()
	result := make([]GameResponse, 0, len(games))
	for _, g := range games {
		result = append(result, gameToResponse(g))
	}
	c.JSON(http.StatusOK, result)
}

func GetGame(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	game := repository.GetGameByID(id)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	c.JSON(http.StatusOK, gameToResponse(game))
}

func UpdateGame(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	found := repository.UpdateGame(id, func(g *model.Game) {
		g.WhitePlayer.Name = req.WhitePlayerName
		g.BlackPlayer.Name = req.BlackPlayerName
		g.Status = model.GameStatus(req.Status)
	})

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	game := repository.GetGameByID(id)
	c.JSON(http.StatusOK, gameToResponse(game))
}
