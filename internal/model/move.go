package model

type Position struct {
	Row int
	Col int
}

type Move struct {
	From   Position
	To     Position
	Player *Player
	Piece  string
}

func NewMove(fromRow, fromCol, toRow, toCol int, player *Player, piece string) *Move {
	return &Move{
		From: Position{
			Row: fromRow,
			Col: fromCol,
		},
		To: Position{
			Row: toRow,
			Col: toCol,
		},
		Player: player,
		Piece:  piece,
	}
}

func (m *Move) IsValid(board *Board) bool {
	return board.IsValidPosition(m.From.Row, m.From.Col) &&
		board.IsValidPosition(m.To.Row, m.To.Col)
}

func (m *Move) GetNotation() string {
	cols := "abcdefgh"
	if m.From.Col < len(cols) && m.To.Col < len(cols) {
		return string(cols[m.From.Col]) + string(rune('1'+m.From.Row)) +
			"-" +
			string(cols[m.To.Col]) + string(rune('1'+m.To.Row))
	}
	return ""
}
