package service

import (
	"math/rand"

	"github.com/imyakin/go_hw/internal/model"
)

func GenerateEntities(store func(model.GameEntity)) {
	switch rand.Intn(4) {
	case 0:
		size := rand.Intn(8) + 1
		store(model.NewBoard(size))
	case 1:
		store(model.NewGame("RandomWhite", "RandomBlack", rand.Intn(8)+1))
	case 2:
		fromRow := rand.Intn(8)
		fromCol := rand.Intn(8)
		toRow := rand.Intn(8)
		toCol := rand.Intn(8)
		player := model.NewPlayer("RandomPlayer", model.White)
		store(model.NewMove(fromRow, fromCol, toRow, toCol, player, "♙"))
	case 3:
		color := model.White
		if rand.Intn(2) == 1 {
			color = model.Black
		}
		store(model.NewPlayer("RandomPlayer", color))
	}
}
