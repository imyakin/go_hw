package main

import "fmt"

func main() {
	size, _, _ := startGame()

	// Calculate row number width for proper alignment
	rowNumberWidth := len(fmt.Sprintf("%d", size))
	// Build column header with letters A, B, C...
	columnHeader := makeColumnHeader(size, rowNumberWidth)
	// Build board with numbers
	board := makeBoard(size, rowNumberWidth)

	fmt.Print(columnHeader + board)
}

func startGame() (int, string, string) {
	var player1, player2 string
	var size int

	fmt.Print("Введите размер доски: ")
	fmt.Scan(&size)
	if size <= 0 {
		fmt.Println("Ошибка: размер доски должен быть больше 0")
		return 0, "", ""
	}
	fmt.Print("Введите имя игрока 1: ")
	fmt.Scan(&player1)
	fmt.Print("Введите имя игрока 2: ")
	fmt.Scan(&player2)

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

func makeBoard(size, rowNumberWidth int) string {
	var board string
	for i := 0; i < size; i++ {
		// Calculate display row number (bottom-to-top, like chess)
		displayRowNum := size - i
		rowLabel := fmt.Sprintf("%*d ", rowNumberWidth, displayRowNum)

		board += rowLabel
		for j := 0; j < size; j++ {
			if (i+j)%2 == 0 {
				board += " "
			} else {
				board += "#"
			}
		}
		board += "\n"
	}
	return board
}
