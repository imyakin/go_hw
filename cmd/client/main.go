package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/render"
)

type GameResponse struct {
	ID                 int        `json:"id"`
	WhitePlayerName    string     `json:"white_player_name"`
	BlackPlayerName    string     `json:"black_player_name"`
	BoardSize          int        `json:"board_size"`
	Status             string     `json:"status"`
	CurrentPlayerColor string     `json:"current_player_color"`
	WinnerColor        string     `json:"winner_color,omitempty"`
	Cells              [][]string `json:"cells"`
}

type MoveResponse struct {
	ID       int    `json:"id"`
	GameID   int    `json:"game_id"`
	FromRow  int    `json:"from_row"`
	FromCol  int    `json:"from_col"`
	ToRow    int    `json:"to_row"`
	ToCol    int    `json:"to_col"`
	Piece    string `json:"piece"`
	Notation string `json:"notation"`
	Player   string `json:"player,omitempty"`
	Error    string `json:"error,omitempty"`
}

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "адрес сервера")
	gameID := flag.Int("game", 0, "ID игры для подключения")
	flag.Parse()

	fmt.Println("=== Шахматный клиент ===")
	fmt.Printf("Сервер: %s\n\n", *serverAddr)

	if *gameID == 0 {
		listGames(*serverAddr)
		fmt.Print("Введите номер игрового поля (ID игры): ")
		fmt.Scan(gameID)
	}

	game := fetchGame(*serverAddr, *gameID)
	if game == nil {
		fmt.Println("Игра не найдена.")
		return
	}

	fmt.Printf("\nПодключено к игре #%d: %s vs %s\n", game.ID, game.WhitePlayerName, game.BlackPlayerName)
	displayGameBoard(game)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		game = fetchGame(*serverAddr, *gameID)
		if game == nil {
			fmt.Println("Ошибка: не удалось получить состояние игры")
			time.Sleep(2 * time.Second)
			continue
		}

		displayGameBoard(game)

		if game.Status == "finished" {
			fmt.Println("\n=== Игра завершена! ===")
			if game.WinnerColor != "" {
				fmt.Printf("Победитель: %s\n", game.WinnerColor)
			}
			break
		}

		fmt.Printf("\nХод: %s\n", game.CurrentPlayerColor)
		fmt.Print("Введите ход (формат e2-e4) или 'exit': ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "exit" || input == "quit" {
			fmt.Println("Выход.")
			break
		}

		if input == "" {
			continue
		}

		resp := sendMove(*serverAddr, *gameID, input)
		if resp.Error != "" {
			fmt.Printf("Ошибка: %s\n", resp.Error)
		} else {
			fmt.Printf("Ход выполнен: %s (ID=%d)\n", resp.Notation, resp.ID)
		}
	}
}

func listGames(serverAddr string) {
	resp, err := http.Get(serverAddr + "/api/games")
	if err != nil {
		fmt.Printf("Ошибка подключения к серверу: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var games []GameResponse
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		fmt.Printf("Ошибка чтения ответа: %v\n", err)
		return
	}

	if len(games) == 0 {
		fmt.Println("Нет доступных игр на сервере.")
	} else {
		fmt.Println("Доступные игры:")
		for _, g := range games {
			fmt.Printf("  [%d] %s vs %s (статус: %s)\n", g.ID, g.WhitePlayerName, g.BlackPlayerName, g.Status)
		}
	}
	fmt.Println()
}

func fetchGame(serverAddr string, gameID int) *GameResponse {
	resp, err := http.Get(fmt.Sprintf("%s/api/games/%d", serverAddr, gameID))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var game GameResponse
	if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
		return nil
	}
	return &game
}

func sendMove(serverAddr string, gameID int, notation string) MoveResponse {
	body := fmt.Sprintf(`{"game_id":%d,"notation":"%s"}`, gameID, notation)
	resp, err := http.Post(
		serverAddr+"/api/moves",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		return MoveResponse{Error: fmt.Sprintf("ошибка подключения: %v", err)}
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]string
		json.Unmarshal(data, &errResp)
		return MoveResponse{Error: errResp["error"]}
	}

	var moveResp MoveResponse
	json.Unmarshal(data, &moveResp)
	return moveResp
}

func displayGameBoard(game *GameResponse) {
	// Reconstruct model.Game for render package
	whitePlayer := model.NewPlayer(game.WhitePlayerName, model.White)
	blackPlayer := model.NewPlayer(game.BlackPlayerName, model.Black)

	board := &model.Board{
		Size:  game.BoardSize,
		Cells: game.Cells,
	}

	g := &model.Game{
		ID:          game.ID,
		WhitePlayer: whitePlayer,
		BlackPlayer: blackPlayer,
		Board:       board,
		Status:      model.GameStatus(game.Status),
		Moves:       make([]*model.Move, 0),
	}

	if game.CurrentPlayerColor == string(model.Black) {
		g.CurrentPlayer = blackPlayer
	} else {
		g.CurrentPlayer = whitePlayer
	}

	render.DisplayBoard(g, game.ID)
}
