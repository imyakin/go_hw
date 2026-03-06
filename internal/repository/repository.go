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
		saveBoardsCSV()
		muBoards.Unlock()
		notifySliceChange("boards", "add", fmt.Sprintf("added board %p, saved to CSV", e))
	case *model.Game:
		muGames.Lock()
		games = append(games, e)
		saveGamesCSV()
		muGames.Unlock()
		notifySliceChange("games", "add", fmt.Sprintf("added game %p, saved to CSV", e))
	case *model.Move:
		muMoves.Lock()
		moves = append(moves, e)
		saveMovesCSV()
		muMoves.Unlock()
		notifySliceChange("moves", "add", fmt.Sprintf("added move %s, saved to CSV", e.GetNotation()))
	case *model.Player:
		muPlayers.Lock()
		players = append(players, e)
		savePlayersCSV()
		muPlayers.Unlock()
		notifySliceChange("players", "add", fmt.Sprintf("added player %s, saved to CSV", e.Name))
	}
}

func RemoveGame(game *model.Game) {
	muGames.Lock()
	defer muGames.Unlock()
	for i, g := range games {
		if g == game {
			games = append(games[:i], games[i+1:]...)
			saveGamesCSV()
			notifySliceChange("games", "remove", fmt.Sprintf("removed game %p, saved to CSV", game))
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
