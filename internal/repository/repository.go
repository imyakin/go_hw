package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
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

var boardIDCounter atomic.Int64
var gameIDCounter atomic.Int64
var moveIDCounter atomic.Int64
var playerIDCounter atomic.Int64

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
		e.ID = int(boardIDCounter.Add(1))
		boards = append(boards, e)
		saveBoardsCSV()
		muBoards.Unlock()
		notifySliceChange("boards", "add", fmt.Sprintf("added board id=%d, saved to CSV", e.ID))
	case *model.Game:
		muGames.Lock()
		e.ID = int(gameIDCounter.Add(1))
		games = append(games, e)
		saveGamesCSV()
		muGames.Unlock()
		notifySliceChange("games", "add", fmt.Sprintf("added game id=%d, saved to CSV", e.ID))
	case *model.Move:
		muMoves.Lock()
		e.ID = int(moveIDCounter.Add(1))
		moves = append(moves, e)
		saveMovesCSV()
		muMoves.Unlock()
		notifySliceChange("moves", "add", fmt.Sprintf("added move id=%d %s, saved to CSV", e.ID, e.GetNotation()))
	case *model.Player:
		muPlayers.Lock()
		e.ID = int(playerIDCounter.Add(1))
		players = append(players, e)
		savePlayersCSV()
		muPlayers.Unlock()
		notifySliceChange("players", "add", fmt.Sprintf("added player id=%d %s, saved to CSV", e.ID, e.Name))
	}
}

func RemoveGame(game *model.Game) {
	muGames.Lock()
	defer muGames.Unlock()
	for i, g := range games {
		if g == game {
			games = append(games[:i], games[i+1:]...)
			saveGamesCSV()
			notifySliceChange("games", "remove", fmt.Sprintf("removed game id=%d, saved to CSV", game.ID))
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

// GetByID functions

func GetGameByID(id int) *model.Game {
	muGames.RLock()
	defer muGames.RUnlock()
	for _, g := range games {
		if g.ID == id {
			return g
		}
	}
	return nil
}

func GetBoardByID(id int) *model.Board {
	muBoards.RLock()
	defer muBoards.RUnlock()
	for _, b := range boards {
		if b.ID == id {
			return b
		}
	}
	return nil
}

func GetPlayerByID(id int) *model.Player {
	muPlayers.RLock()
	defer muPlayers.RUnlock()
	for _, p := range players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func GetMoveByID(id int) *model.Move {
	muMoves.RLock()
	defer muMoves.RUnlock()
	for _, m := range moves {
		if m.ID == id {
			return m
		}
	}
	return nil
}

// Update functions -- apply mutation under write lock and save CSV

func UpdateGame(id int, fn func(*model.Game)) bool {
	muGames.Lock()
	defer muGames.Unlock()
	for _, g := range games {
		if g.ID == id {
			fn(g)
			saveGamesCSV()
			return true
		}
	}
	return false
}

func UpdateBoard(id int, fn func(*model.Board)) bool {
	muBoards.Lock()
	defer muBoards.Unlock()
	for _, b := range boards {
		if b.ID == id {
			fn(b)
			saveBoardsCSV()
			return true
		}
	}
	return false
}

func UpdatePlayer(id int, fn func(*model.Player)) bool {
	muPlayers.Lock()
	defer muPlayers.Unlock()
	for _, p := range players {
		if p.ID == id {
			fn(p)
			savePlayersCSV()
			return true
		}
	}
	return false
}

func UpdateMove(id int, fn func(*model.Move)) bool {
	muMoves.Lock()
	defer muMoves.Unlock()
	for _, m := range moves {
		if m.ID == id {
			fn(m)
			saveMovesCSV()
			return true
		}
	}
	return false
}

func syncCounters() {
	for _, g := range games {
		if int64(g.ID) > gameIDCounter.Load() {
			gameIDCounter.Store(int64(g.ID))
		}
	}
	for _, b := range boards {
		if int64(b.ID) > boardIDCounter.Load() {
			boardIDCounter.Store(int64(b.ID))
		}
	}
	for _, p := range players {
		if int64(p.ID) > playerIDCounter.Load() {
			playerIDCounter.Store(int64(p.ID))
		}
	}
	for _, m := range moves {
		if int64(m.ID) > moveIDCounter.Load() {
			moveIDCounter.Store(int64(m.ID))
		}
	}
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
