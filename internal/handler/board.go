package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/repository"
)

type CreateBoardRequest struct {
	Size int `json:"size" binding:"required,min=1"`
}

type UpdateBoardRequest struct {
	Size  int        `json:"size" binding:"required,min=1"`
	Cells [][]string `json:"cells" binding:"required"`
}

type BoardResponse struct {
	ID    int        `json:"id"`
	Size  int        `json:"size"`
	Cells [][]string `json:"cells"`
}

func boardToResponse(b *model.Board) BoardResponse {
	return BoardResponse{
		ID:    b.ID,
		Size:  b.Size,
		Cells: b.Cells,
	}
}

func CreateBoard(c *gin.Context) {
	var req CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	board := model.NewBoard(req.Size)
	repository.Store(board)

	c.JSON(http.StatusCreated, boardToResponse(board))
}

func ListBoards(c *gin.Context) {
	boards := repository.GetBoards()
	result := make([]BoardResponse, 0, len(boards))
	for _, b := range boards {
		result = append(result, boardToResponse(b))
	}
	c.JSON(http.StatusOK, result)
}

func GetBoard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	board := repository.GetBoardByID(id)
	if board == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}

	c.JSON(http.StatusOK, boardToResponse(board))
}

func UpdateBoard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	found := repository.UpdateBoard(id, func(b *model.Board) {
		b.Size = req.Size
		b.Cells = req.Cells
	})

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}

	board := repository.GetBoardByID(id)
	c.JSON(http.StatusOK, boardToResponse(board))
}
