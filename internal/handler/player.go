package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/repository"
)

type CreatePlayerRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,oneof=white black"`
}

type UpdatePlayerRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,oneof=white black"`
}

type PlayerResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Symbol string `json:"symbol"`
}

func playerToResponse(p *model.Player) PlayerResponse {
	return PlayerResponse{
		ID:     p.ID,
		Name:   p.Name,
		Color:  string(p.Color),
		Symbol: p.Symbol,
	}
}

// CreatePlayer creates a new player
// @Summary Create a new player
// @Description Creates a new chess player with name and color
// @Tags players
// @Accept json
// @Produce json
// @Param input body CreatePlayerRequest true "Player creation parameters"
// @Success 201 {object} PlayerResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/players [post]
func CreatePlayer(c *gin.Context) {
	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := model.NewPlayer(req.Name, model.PlayerColor(req.Color))
	repository.Store(player)

	c.JSON(http.StatusCreated, playerToResponse(player))
}

// ListPlayers returns all players
// @Summary List all players
// @Description Returns a list of all chess players
// @Tags players
// @Produce json
// @Success 200 {array} PlayerResponse
// @Router /api/players [get]
func ListPlayers(c *gin.Context) {
	players := repository.GetPlayers()
	result := make([]PlayerResponse, 0, len(players))
	for _, p := range players {
		result = append(result, playerToResponse(p))
	}
	c.JSON(http.StatusOK, result)
}

// GetPlayer returns a player by ID
// @Summary Get a player by ID
// @Description Returns a single chess player by its ID
// @Tags players
// @Produce json
// @Param id path int true "Player ID"
// @Success 200 {object} PlayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/players/{id} [get]
func GetPlayer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	player := repository.GetPlayerByID(id)
	if player == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	c.JSON(http.StatusOK, playerToResponse(player))
}

// UpdatePlayer updates an existing player
// @Summary Update a player
// @Description Updates an existing chess player by ID
// @Tags players
// @Accept json
// @Produce json
// @Param id path int true "Player ID"
// @Param input body UpdatePlayerRequest true "Player update parameters"
// @Success 200 {object} PlayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/players/{id} [put]
func UpdatePlayer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	found := repository.UpdatePlayer(id, func(p *model.Player) {
		p.Name = req.Name
		p.Color = model.PlayerColor(req.Color)
		if req.Color == "black" {
			p.Symbol = "♚"
		} else {
			p.Symbol = "♔"
		}
	})

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	player := repository.GetPlayerByID(id)
	c.JSON(http.StatusOK, playerToResponse(player))
}
