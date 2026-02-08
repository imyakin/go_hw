package repository

import (
	"fmt"

	"github.com/imyakin/go_hw/internal/model"
)

var boards []*model.Board
var games []*model.Game
var moves []*model.Move
var players []*model.Player

func Store(entity model.GameEntity) {
	switch e := entity.(type) {
	case *model.Board:
		boards = append(boards, e)
	case *model.Game:
		games = append(games, e)
	case *model.Move:
		moves = append(moves, e)
	case *model.Player:
		players = append(players, e)
	}
}

func PrintStats() {
	fmt.Println("=== Repository Stats ===")
	fmt.Printf("Boards:  %d\n", len(boards))
	fmt.Printf("Games:   %d\n", len(games))
	fmt.Printf("Moves:   %d\n", len(moves))
	fmt.Printf("Players: %d\n", len(players))
}
