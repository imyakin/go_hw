package render

import (
	"fmt"
	"strings"

	"github.com/imyakin/go_hw/internal/model"
)

var WhitePieces = map[string]string{
	"king":   "♔",
	"queen":  "♕",
	"rook":   "♖",
	"bishop": "♗",
	"knight": "♘",
	"pawn":   "♙",
}

var BlackPieces = map[string]string{
	"king":   "♚",
	"queen":  "♛",
	"rook":   "♜",
	"bishop": "♝",
	"knight": "♞",
	"pawn":   "♟",
}

func PlacePieces(board *model.Board, size int) {
	if size < 8 {
		return
	}

	board.SetCell(0, 0, BlackPieces["rook"])
	board.SetCell(0, 1, BlackPieces["knight"])
	board.SetCell(0, 2, BlackPieces["bishop"])
	board.SetCell(0, 3, BlackPieces["queen"])
	board.SetCell(0, 4, BlackPieces["king"])
	board.SetCell(0, 5, BlackPieces["bishop"])
	board.SetCell(0, 6, BlackPieces["knight"])
	board.SetCell(0, 7, BlackPieces["rook"])

	for col := 0; col < 8; col++ {
		board.SetCell(1, col, BlackPieces["pawn"])
	}

	for col := 0; col < 8; col++ {
		board.SetCell(size-2, col, WhitePieces["pawn"])
	}

	board.SetCell(size-1, 0, WhitePieces["rook"])
	board.SetCell(size-1, 1, WhitePieces["knight"])
	board.SetCell(size-1, 2, WhitePieces["bishop"])
	board.SetCell(size-1, 3, WhitePieces["queen"])
	board.SetCell(size-1, 4, WhitePieces["king"])
	board.SetCell(size-1, 5, WhitePieces["bishop"])
	board.SetCell(size-1, 6, WhitePieces["knight"])
	board.SetCell(size-1, 7, WhitePieces["rook"])
}

func MakeColumnHeader(size, rowNumberWidth int) string {
	columnHeader := ""
	for i := 0; i < rowNumberWidth+1; i++ {
		columnHeader += " "
	}
	for j := 0; j < size; j++ {
		columnHeader += string(rune('A' + j%26))
	}
	columnHeader += "\n"
	return columnHeader
}

func DisplayBoard(game *model.Game, index int) {
	game.Mu.RLock()
	defer game.Mu.RUnlock()
	board := game.Board
	size := board.Size
	rowNumberWidth := len(fmt.Sprintf("%d", size))

	fmt.Printf("\nИгра #%d\n", index)
	fmt.Printf("Время хода: %s %v | %s %v\n",
		game.WhitePlayer.GetDisplayName(),
		game.LastWhiteTime,
		game.BlackPlayer.GetDisplayName(),
		game.LastBlackTime,
	)

	fmt.Print(MakeColumnHeader(size, rowNumberWidth))

	player1Row := 1
	player2Row := size

	for i := 0; i < size; i++ {
		displayRowNum := size - i
		rowLabel := fmt.Sprintf("%*d ", rowNumberWidth, displayRowNum)

		fmt.Print(rowLabel)
		for j := 0; j < size; j++ {
			piece := board.GetCell(i, j)
			if piece != "" {
				fmt.Print(piece)
			} else {
				if (i+j)%2 == 0 {
					fmt.Print(" ")
				} else {
					fmt.Print("#")
				}
			}
		}

		if displayRowNum == player1Row {
			fmt.Print("  " + game.WhitePlayer.GetDisplayName())
		} else if displayRowNum == player2Row {
			fmt.Print("  " + game.BlackPlayer.GetDisplayName())
		}

		fmt.Println()
	}
}

func BoardToString(game *model.Game) string {
	game.Mu.RLock()
	defer game.Mu.RUnlock()
	board := game.Board
	size := board.Size
	rowNumberWidth := len(fmt.Sprintf("%d", size))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Игра #%d (ID=%d)\n", game.ID, game.ID))
	sb.WriteString(fmt.Sprintf("%s vs %s | Статус: %s\n",
		game.WhitePlayer.GetDisplayName(),
		game.BlackPlayer.GetDisplayName(),
		game.Status,
	))

	sb.WriteString(MakeColumnHeader(size, rowNumberWidth))

	for i := 0; i < size; i++ {
		displayRowNum := size - i
		sb.WriteString(fmt.Sprintf("%*d ", rowNumberWidth, displayRowNum))
		for j := 0; j < size; j++ {
			piece := board.GetCell(i, j)
			if piece != "" {
				sb.WriteString(piece)
			} else {
				if (i+j)%2 == 0 {
					sb.WriteString(" ")
				} else {
					sb.WriteString("#")
				}
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func IsPlayerPiece(piece string, player *model.Player) bool {
	if player.IsWhite() {
		return piece == "♔" || piece == "♕" || piece == "♖" || piece == "♗" || piece == "♘" || piece == "♙"
	}
	return piece == "♚" || piece == "♛" || piece == "♜" || piece == "♝" || piece == "♞" || piece == "♟"
}

func ParseMove(input string, player *model.Player, board *model.Board) (*model.Move, error) {
	if len(input) < 5 || input[2] != '-' {
		return nil, fmt.Errorf("неверный формат хода. Используйте формат: e2-e4")
	}

	fromCol := ConvertColumnToIndex(input[0])
	fromRow := int(input[1] - '1')
	toCol := ConvertColumnToIndex(input[3])
	toRow := int(input[4] - '1')

	if fromCol < 0 || toCol < 0 {
		return nil, fmt.Errorf("неверная колонка")
	}

	fromRowActual := board.Size - 1 - fromRow
	toRowActual := board.Size - 1 - toRow

	piece := board.GetCell(fromRowActual, fromCol)
	if piece == "" {
		return nil, fmt.Errorf("на клетке %c%d нет фигуры", input[0], fromRow+1)
	}

	return model.NewMove(fromRowActual, fromCol, toRowActual, toCol, player, piece), nil
}

func ApplyMove(board *model.Board, move *model.Move) error {
	piece := board.GetCell(move.From.Row, move.From.Col)
	if piece == "" {
		return fmt.Errorf("нет фигуры на начальной позиции")
	}

	board.SetCell(move.To.Row, move.To.Col, piece)
	board.SetCell(move.From.Row, move.From.Col, "")

	return nil
}

func ConvertColumnToIndex(col byte) int {
	if col >= 'a' && col <= 'z' {
		return int(col - 'a')
	}
	if col >= 'A' && col <= 'Z' {
		return int(col - 'A')
	}
	return -1
}
