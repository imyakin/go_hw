package model

import (
	"sync"
	"time"
)

type GameStatus string

const (
	StatusNotStarted GameStatus = "not_started"
	StatusInProgress GameStatus = "in_progress"
	StatusFinished   GameStatus = "finished"
)

type Game struct {
	WhitePlayer   *Player
	BlackPlayer   *Player
	Board         *Board
	Moves         []*Move
	CurrentPlayer *Player
	Status        GameStatus
	Winner        *Player
	Mu            sync.RWMutex
	LastMoveTime  time.Duration
	LastWhiteTime time.Duration
	LastBlackTime time.Duration
	MoveStartTime time.Time
}

func NewGame(whitePlayerName, blackPlayerName string, boardSize int) *Game {
	whitePlayer := NewPlayer(whitePlayerName, White)
	blackPlayer := NewPlayer(blackPlayerName, Black)
	board := NewBoard(boardSize)

	return &Game{
		WhitePlayer:   whitePlayer,
		BlackPlayer:   blackPlayer,
		Board:         board,
		Moves:         make([]*Move, 0),
		CurrentPlayer: whitePlayer,
		Status:        StatusNotStarted,
	}
}

func (g *Game) Start() {
	g.Status = StatusInProgress
}

func (g *Game) IsInProgress() bool {
	return g.Status == StatusInProgress
}

func (g *Game) IsFinished() bool {
	return g.Status == StatusFinished
}

func (g *Game) MakeMove(move *Move) bool {
	if !g.IsInProgress() {
		return false
	}

	if move.Player != g.CurrentPlayer {
		return false
	}

	if !move.IsValid(g.Board) {
		return false
	}

	g.Moves = append(g.Moves, move)
	g.SwitchPlayer()
	return true
}

func (g *Game) SwitchPlayer() {
	if g.CurrentPlayer == g.WhitePlayer {
		g.CurrentPlayer = g.BlackPlayer
	} else {
		g.CurrentPlayer = g.WhitePlayer
	}
}

func (g *Game) Finish() {
	g.Status = StatusFinished
}

func (g *Game) GetMoveHistory() []*Move {
	return g.Moves
}

func (g *Game) GetMoveCount() int {
	return len(g.Moves)
}

func (g *Game) EntityType() string {
	return "Game"
}

func (g *Game) Resign() {
	if g.CurrentPlayer == g.WhitePlayer {
		g.Winner = g.BlackPlayer
	} else {
		g.Winner = g.WhitePlayer
	}
	g.Finish()
}
