package main

import "fmt"

func main() {
	var size int
	fmt.Print("Введите размер доски: ")
	fmt.Scan(&size)

	if size <= 0 {
		fmt.Println("Ошибка: размер доски должен быть больше 0")
		return
	}

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
