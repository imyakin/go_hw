package main

import (
	"fmt"

	"github.com/imyakin/go_hw/internal/model"
)

var whitePieces = map[string]string{
	"king":   "♔",
	"queen":  "♕",
	"rook":   "♖",
	"bishop": "♗",
	"knight": "♘",
	"pawn":   "♙",
}

var blackPieces = map[string]string{
	"king":   "♚",
	"queen":  "♛",
	"rook":   "♜",
	"bishop": "♝",
	"knight": "♞",
	"pawn":   "♟",
}

func main() {
	size, player1, player2 := startGame()

	// Calculate row number width for proper alignment
	rowNumberWidth := len(fmt.Sprintf("%d", size))
	// Build column header with letters A, B, C...
	columnHeader := makeColumnHeader(size, rowNumberWidth)
	// Build board with numbers and player names
	board := makeBoard(size, rowNumberWidth, player1, player2)

	fmt.Print(columnHeader + board)
}

func startGame() (int, *model.Player, *model.Player) {
	var player1Name, player2Name string
	var size int

	fmt.Print("Введите размер доски: ")
	fmt.Scan(&size)
	if size <= 0 {
		fmt.Println("Ошибка: размер доски должен быть больше 0")
		return 0, nil, nil
	}
	fmt.Print("Введите имя игрока 1: ")
	fmt.Scan(&player1Name)
	fmt.Print("Введите имя игрока 2: ")
	fmt.Scan(&player2Name)

	player1 := model.NewPlayer(player1Name, model.White)
	player2 := model.NewPlayer(player2Name, model.Black)

	return size, player1, player2
}

func makeColumnHeader(size, rowNumberWidth int) string {
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

func makeBoard(size, rowNumberWidth int, player1, player2 *model.Player) string {
	var board string

	player1Row := 1    // White pieces (bottom)
	player2Row := size // Black pieces (top)

	for i := 0; i < size; i++ {
		// Calculate display row number (bottom-to-top, like chess)
		displayRowNum := size - i
		rowLabel := fmt.Sprintf("%*d ", rowNumberWidth, displayRowNum)

		board += rowLabel
		for j := 0; j < size; j++ {
			piece := getPieceAt(displayRowNum, j, size)
			if piece != "" {
				board += piece
			} else {
				if (i+j)%2 == 0 {
					board += " "
				} else {
					board += "#"
				}
			}
		}

		// Add player names on the side
		if displayRowNum == player1Row {
			board += "  " + player1.GetDisplayName()
		} else if displayRowNum == player2Row {
			board += "  " + player2.GetDisplayName()
		}

		board += "\n"
	}
	return board
}

func getPieceAt(row, col, size int) string {
	// Only place pieces on 8x8 board or larger
	if size < 8 {
		return ""
	}

	// Black pieces on row 8
	if row == size {
		switch col {
		case 0, 7:
			return blackPieces["rook"]
		case 1, 6:
			return blackPieces["knight"]
		case 2, 5:
			return blackPieces["bishop"]
		case 3:
			return blackPieces["queen"]
		case 4:
			return blackPieces["king"]
		}
	}

	// Black pawns on row 7
	if row == size-1 && col < 8 {
		return blackPieces["pawn"]
	}

	// White pawns on row 2
	if row == 2 && col < 8 {
		return whitePieces["pawn"]
	}

	// White pieces on row 1
	if row == 1 {
		switch col {
		case 0, 7:
			return whitePieces["rook"]
		case 1, 6:
			return whitePieces["knight"]
		case 2, 5:
			return whitePieces["bishop"]
		case 3:
			return whitePieces["queen"]
		case 4:
			return whitePieces["king"]
		}
	}

	return ""
}
