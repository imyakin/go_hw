package model

type PlayerColor string

const (
	White PlayerColor = "white"
	Black PlayerColor = "black"
)

type Player struct {
	Name   string
	Color  PlayerColor
	Symbol string // King symbol for the player (♔ or ♚)
}

func NewPlayer(name string, color PlayerColor) *Player {
	symbol := "♔" // White king by default
	if color == Black {
		symbol = "♚"
	}

	return &Player{
		Name:   name,
		Color:  color,
		Symbol: symbol,
	}
}

func (p *Player) IsWhite() bool {
	return p.Color == White
}

func (p *Player) IsBlack() bool {
	return p.Color == Black
}

func (p *Player) GetDisplayName() string {
	colorName := "белые"
	if p.Color == Black {
		colorName = "черные"
	}
	return p.Name + " (" + colorName + " " + p.Symbol + ")"
}
