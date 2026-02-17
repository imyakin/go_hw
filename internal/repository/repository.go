package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/imyakin/go_hw/internal/model"
)

var boards []*model.Board
var games []*model.Game
var moves []*model.Move
var players []*model.Player
var muBoards sync.RWMutex
var muGames sync.RWMutex
var muMoves sync.RWMutex
var muPlayers sync.RWMutex

type SliceChange struct {
	SliceType string
	Operation string
	Timestamp time.Time
	Details   string
}

var SliceChangeChan = make(chan SliceChange, 128)

func Store(entity model.GameEntity) {
	switch e := entity.(type) {
	case *model.Board:
		muBoards.Lock()
		boards = append(boards, e)
		muBoards.Unlock()
	case *model.Game:
		muGames.Lock()
		games = append(games, e)
		muGames.Unlock()
	case *model.Move:
		muMoves.Lock()
		moves = append(moves, e)
		muMoves.Unlock()
	case *model.Player:
		muPlayers.Lock()
		players = append(players, e)
		muPlayers.Unlock()
	}
}

func RemoveGame(game *model.Game) {
	muGames.Lock()
	defer muGames.Unlock()
	for i, g := range games {
		if g == game {
			games = append(games[:i], games[i+1:]...)
			notifySliceChange("games", "remove", fmt.Sprintf("removed game %p", game))
			return
		}
	}
}

func GetBoards() []*model.Board {
	muBoards.RLock()
	defer muBoards.RUnlock()
	copied := make([]*model.Board, len(boards))
	copy(copied, boards)
	return copied
}

func GetGames() []*model.Game {
	muGames.RLock()
	defer muGames.RUnlock()
	copied := make([]*model.Game, len(games))
	copy(copied, games)
	return copied
}

func GetMoves() []*model.Move {
	muMoves.RLock()
	defer muMoves.RUnlock()
	copied := make([]*model.Move, len(moves))
	copy(copied, moves)
	return copied
}

func GetPlayers() []*model.Player {
	muPlayers.RLock()
	defer muPlayers.RUnlock()
	copied := make([]*model.Player, len(players))
	copy(copied, players)
	return copied
}

func notifySliceChange(sliceType, operation, details string) {
	select {
	case SliceChangeChan <- SliceChange{
		SliceType: sliceType,
		Operation: operation,
		Timestamp: time.Now(),
		Details:   details,
	}:
	default:
	}
}

func LogSliceChange(sliceType, operation, details string) {
	notifySliceChange(sliceType, operation, details)
}

func PrintStats() {
	fmt.Println("=== Repository Stats ===")
	fmt.Printf("Boards:  %d\n", len(GetBoards()))
	fmt.Printf("Games:   %d\n", len(GetGames()))
	fmt.Printf("Moves:   %d\n", len(GetMoves()))
	fmt.Printf("Players: %d\n", len(GetPlayers()))
}
