package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	chessv1 "github.com/imyakin/go_hw/proto/chess/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC адрес сервера")
	cmd := flag.String("cmd", "demo", "команда: list | create | move | demo")
	gameID := flag.Int("game", 0, "ID игры (для move)")
	notation := flag.String("notation", "e2-e4", "нотация хода")
	flag.Parse()

	conn, err := grpc.NewClient(
		*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gameClient := chessv1.NewGameServiceClient(conn)
	moveClient := chessv1.NewMoveServiceClient(conn)

	switch *cmd {
	case "list":
		listGames(ctx, gameClient)

	case "create":
		createGame(ctx, gameClient)

	case "move":
		if *gameID == 0 {
			log.Fatal("укажи -game <id>")
		}
		createMove(ctx, moveClient, int32(*gameID), *notation)

	case "demo":
		fmt.Println("=== ListGames (до создания) ===")
		listGames(ctx, gameClient)

		fmt.Println("\n=== CreateGame ===")
		g := createGame(ctx, gameClient)

		fmt.Println("\n=== ListGames (после создания) ===")
		listGames(ctx, gameClient)

		fmt.Println("\n=== CreateMove e2-e4 ===")
		createMove(ctx, moveClient, g.GetId(), "e2-e4")

	default:
		log.Fatalf("неизвестная команда: %s", *cmd)
	}
}

func listGames(ctx context.Context, c chessv1.GameServiceClient) {
	resp, err := c.ListGames(ctx, &chessv1.ListRequest{})
	if err != nil {
		log.Fatalf("ListGames: %v", err)
	}
	if len(resp.GetItems()) == 0 {
		fmt.Println("игр нет")
		return
	}
	for _, g := range resp.GetItems() {
		fmt.Printf("[%d] %s vs %s | status=%s | board=%d\n",
			g.GetId(), g.GetWhitePlayerName(), g.GetBlackPlayerName(),
			g.GetStatus(), g.GetBoardSize())
	}
}

func createGame(ctx context.Context, c chessv1.GameServiceClient) *chessv1.Game {
	resp, err := c.CreateGame(ctx, &chessv1.CreateGameRequest{
		WhitePlayerName: "Alice",
		BlackPlayerName: "Bob",
		BoardSize:       8,
	})
	if err != nil {
		log.Fatalf("CreateGame: %v", err)
	}
	fmt.Printf("создана игра id=%d: %s vs %s\n",
		resp.GetId(), resp.GetWhitePlayerName(), resp.GetBlackPlayerName())
	return resp
}

func createMove(ctx context.Context, c chessv1.MoveServiceClient, gameID int32, notation string) {
	resp, err := c.CreateMove(ctx, &chessv1.CreateMoveRequest{
		GameId:   gameID,
		Notation: notation,
	})
	if err != nil {
		log.Fatalf("CreateMove: %v", err)
	}
	fmt.Printf("ход id=%d game_id=%d notation=%s piece=%s\n",
		resp.GetId(), resp.GetGameId(), resp.GetNotation(), resp.GetPiece())
}
