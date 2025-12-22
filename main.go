package main

import "fmt"

func main() {
	size := 8

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
