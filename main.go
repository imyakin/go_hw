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
	game := startGame()
	if game == nil {
		return
	}

	placePieces(game.Board, game.Board.Size)
	game.Start()
	displayBoard(game)

	// Game loop
	gameLoop(game)
}

func startGame() *model.Game {
	var player1Name, player2Name string
	var size int

	fmt.Print("Введите размер доски: ")
	fmt.Scan(&size)
	if size <= 0 {
		fmt.Println("Ошибка: размер доски должен быть больше 0")
		return nil
	}
	fmt.Print("Введите имя игрока 1: ")
	fmt.Scan(&player1Name)
	fmt.Print("Введите имя игрока 2: ")
	fmt.Scan(&player2Name)

	game := model.NewGame(player1Name, player2Name, size)
	return game
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

func displayBoard(game *model.Game) {
	board := game.Board
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
			fmt.Print("  " + game.WhitePlayer.GetDisplayName())
		} else if displayRowNum == player2Row {
			fmt.Print("  " + game.BlackPlayer.GetDisplayName())
		}

		fmt.Println()
	}
}

func gameLoop(game *model.Game) {
	for game.IsInProgress() {
		fmt.Printf("\n%s, ваш ход (формат: e2-e4 или 'exit' для выхода): ", game.CurrentPlayer.GetDisplayName())

		var input string
		fmt.Scan(&input)

		if input == "exit" || input == "quit" {
			game.Finish()
			fmt.Println("Игра завершена!")
			fmt.Printf("Всего ходов: %d\n", game.GetMoveCount())
			break
		}

		move, err := parseMove(input, game.CurrentPlayer, game.Board)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		if !move.IsValid(game.Board) {
			fmt.Println("Ошибка: неверные координаты хода")
			continue
		}

		if err := applyMove(game.Board, move); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		// Add move to game history and switch player
		game.MakeMove(move)

		fmt.Println()
		displayBoard(game)
	}
}

func parseMove(input string, player *model.Player, board *model.Board) (*model.Move, error) {
	if len(input) < 5 || input[2] != '-' {
		return nil, fmt.Errorf("неверный формат хода. Используйте формат: e2-e4")
	}

	fromCol := convertColumnToIndex(input[0])
	fromRow := int(input[1] - '1')
	toCol := convertColumnToIndex(input[3])
	toRow := int(input[4] - '1')

	if fromCol < 0 || toCol < 0 {
		return nil, fmt.Errorf("неверная колонка")
	}

	// Convert display row to actual board row (inverted)
	fromRowActual := board.Size - 1 - fromRow
	toRowActual := board.Size - 1 - toRow

	piece := board.GetCell(fromRowActual, fromCol)
	if piece == "" {
		return nil, fmt.Errorf("на клетке %c%d нет фигуры", input[0], fromRow+1)
	}

	return model.NewMove(fromRowActual, fromCol, toRowActual, toCol, player, piece), nil
}

func applyMove(board *model.Board, move *model.Move) error {
	piece := board.GetCell(move.From.Row, move.From.Col)
	if piece == "" {
		return fmt.Errorf("нет фигуры на начальной позиции")
	}

	board.SetCell(move.To.Row, move.To.Col, piece)
	board.SetCell(move.From.Row, move.From.Col, "")

	return nil
}

func convertColumnToIndex(col byte) int {
	if col >= 'a' && col <= 'z' {
		return int(col - 'a')
	}
	if col >= 'A' && col <= 'Z' {
		return int(col - 'A')
	}
	return -1
}
