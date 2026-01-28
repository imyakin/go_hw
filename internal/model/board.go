package model

type Board struct {
	Size  int
	Cells [][]string
}

func NewBoard(size int) *Board {
	cells := make([][]string, size)
	for i := range cells {
		cells[i] = make([]string, size)
	}

	return &Board{
		Size:  size,
		Cells: cells,
	}
}

func (b *Board) GetCell(row, col int) string {
	if row < 0 || row >= b.Size || col < 0 || col >= b.Size {
		return ""
	}
	return b.Cells[row][col]
}

func (b *Board) SetCell(row, col int, piece string) {
	if row >= 0 && row < b.Size && col >= 0 && col < b.Size {
		b.Cells[row][col] = piece
	}
}

func (b *Board) IsValidPosition(row, col int) bool {
	return row >= 0 && row < b.Size && col >= 0 && col < b.Size
}

func (b *Board) EntityType() string {
	return "Board"
}
