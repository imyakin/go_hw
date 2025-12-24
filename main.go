package main

import "fmt"

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

func main() {
	size, _, _ := startGame()

	var board string

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if (i+j)%2 == 0 {
				board += " "
			} else {
				board += "#"
			}
		}
		board += "\n"
	}

	fmt.Print(board)
}
