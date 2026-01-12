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

	board := model.NewBoard(size)
	placePieces(board, size)
	displayBoard(board, player1, player2)
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

func placePieces(board *model.Board, size int) {
	// Only place pieces on 8x8 board or larger
	if size < 8 {
		return
	}

	// Black pieces on row index 0 (displayed as row 8)
	board.SetCell(0, 0, blackPieces["rook"])
	board.SetCell(0, 1, blackPieces["knight"])
	board.SetCell(0, 2, blackPieces["bishop"])
	board.SetCell(0, 3, blackPieces["queen"])
	board.SetCell(0, 4, blackPieces["king"])
	board.SetCell(0, 5, blackPieces["bishop"])
	board.SetCell(0, 6, blackPieces["knight"])
	board.SetCell(0, 7, blackPieces["rook"])

	// Black pawns on row index 1 (displayed as row 7)
	for col := 0; col < 8; col++ {
		board.SetCell(1, col, blackPieces["pawn"])
	}

	// White pawns on row index size-2 (displayed as row 2)
	for col := 0; col < 8; col++ {
		board.SetCell(size-2, col, whitePieces["pawn"])
	}

	// White pieces on row index size-1 (displayed as row 1)
	board.SetCell(size-1, 0, whitePieces["rook"])
	board.SetCell(size-1, 1, whitePieces["knight"])
	board.SetCell(size-1, 2, whitePieces["bishop"])
	board.SetCell(size-1, 3, whitePieces["queen"])
	board.SetCell(size-1, 4, whitePieces["king"])
	board.SetCell(size-1, 5, whitePieces["bishop"])
	board.SetCell(size-1, 6, whitePieces["knight"])
	board.SetCell(size-1, 7, whitePieces["rook"])
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

func displayBoard(board *model.Board, player1, player2 *model.Player) {
	size := board.Size
	rowNumberWidth := len(fmt.Sprintf("%d", size))

	// Print column header
	fmt.Print(makeColumnHeader(size, rowNumberWidth))

	player1Row := 1    // White pieces (bottom)
	player2Row := size // Black pieces (top)

	for i := 0; i < size; i++ {
		// Calculate display row number (bottom-to-top, like chess)
		displayRowNum := size - i
		rowLabel := fmt.Sprintf("%*d ", rowNumberWidth, displayRowNum)

		fmt.Print(rowLabel)
		for j := 0; j < size; j++ {
			// Get cell content using GetCell method
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

		// Add player names on the side
		if displayRowNum == player1Row {
			fmt.Print("  " + player1.GetDisplayName())
		} else if displayRowNum == player2Row {
			fmt.Print("  " + player2.GetDisplayName())
		}

		fmt.Println()
	}
}
