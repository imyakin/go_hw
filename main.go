package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/render"
	"github.com/imyakin/go_hw/internal/repository"
)

type GameManager struct {
	games []*model.Game
	mu    sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make([]*model.Game, 0),
	}
}

func (m *GameManager) AddGame(game *model.Game) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.games = append(m.games, game)
	repository.LogSliceChange("games", "add", fmt.Sprintf("added game %p", game))
}

func (m *GameManager) GetGames() []*model.Game {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]*model.Game, len(m.games))
	copy(copied, m.games)
	return copied
}

func (m *GameManager) RemoveFinishedGames() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.games) == 0 {
		return
	}
	remaining := make([]*model.Game, 0, len(m.games))
	for _, game := range m.games {
		game.Mu.RLock()
		finished := game.IsFinished()
		game.Mu.RUnlock()
		if finished {
			repository.RemoveGame(game)
			continue
		}
		remaining = append(remaining, game)
	}
	m.games = remaining
}

func (m *GameManager) GetGameCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.games)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := repository.LoadAll(); err != nil {
		fmt.Printf("Предупреждение: ошибка загрузки данных из CSV: %v\n", err)
	} else {
		fmt.Println("Данные из предыдущих сессий загружены из CSV файлов.")
		repository.PrintStats()
	}

	games := startGames()
	if len(games) == 0 {
		return
	}

	manager := NewGameManager()
	for _, game := range games {
		render.PlacePieces(game.Board, game.Board.Size)
		game.Start()
		manager.AddGame(game)
		repository.Store(game)
		repository.Store(game.Board)
		repository.Store(game.WhitePlayer)
		repository.Store(game.BlackPlayer)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sliceLogger(ctx)
	}()

	if manager.GetGameCount() == 1 {
		game := manager.GetGames()[0]
		render.DisplayBoard(game, 1)
		gameLoop(ctx, game)
	} else {
		updateCh := make(chan []*model.Game, 16)
		wg.Add(2)
		go func() {
			defer wg.Done()
			gameSimulator(ctx, manager, updateCh)
		}()
		go func() {
			defer wg.Done()
			boardRenderer(ctx, manager, updateCh)
		}()

		for manager.GetGameCount() > 0 {
			select {
			case <-ctx.Done():
				goto waitWorkers
			default:
				time.Sleep(200 * time.Millisecond)
			}
		}
	waitWorkers:
	}

	exitedBySignal := (ctx.Err() != nil)
	stop()
	wg.Wait()

	if !exitedBySignal {
		repository.PrintStats()
	}
}

func startGames() []*model.Game {
	var player1Name, player2Name string
	var size int
	var gameCount int

	fmt.Print("Введите количество досок: ")
	fmt.Scan(&gameCount)
	if gameCount <= 0 {
		fmt.Println("Ошибка: количество досок должно быть больше 0")
		return nil
	}

	games := make([]*model.Game, 0, gameCount)
	for i := 0; i < gameCount; i++ {
		fmt.Printf("Доска %d\n", i+1)
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
		games = append(games, game)
	}

	return games
}

func gameLoop(ctx context.Context, game *model.Game) {
	inputCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputCh <- scanner.Text()
		}
		close(inputCh)
	}()

	for game.IsInProgress() {
		fmt.Printf("\n%s, ваш ход (формат: e2-e4 или 'exit' для выхода или 'Автоход' или 'Сдался'): ", game.CurrentPlayer.GetDisplayName())

		var input string
		select {
		case <-ctx.Done():
			return
		case line, ok := <-inputCh:
			if !ok {
				return
			}
			input = line
		}

		startTime := time.Now()

		if input == "exit" || input == "quit" {
			game.Finish()
			fmt.Println("Игра завершена!")
			fmt.Printf("Всего ходов: %d\n", game.GetMoveCount())
			break
		}

		if strings.EqualFold(input, "Сдался") {
			resigned := game.CurrentPlayer
			game.Resign()
			fmt.Printf("%s сдался! Победил %s!\n", resigned.GetDisplayName(), game.Winner.GetDisplayName())
			break
		}

		if strings.EqualFold(input, "Автоход") {
			fmt.Print("Сколько автоходов сделать: ")
			var count int
			select {
			case <-ctx.Done():
				return
			case line, ok := <-inputCh:
				if !ok {
					return
				}
				if _, err := fmt.Sscanf(line, "%d", &count); err != nil || count <= 0 {
					fmt.Println("Ошибка: введите положительное число")
					continue
				}
			}
			for i := 0; i < count; i++ {
				if !game.IsInProgress() {
					break
				}
				select {
				case <-ctx.Done():
					return
				default:
				}
				duration, notation, mover, err := autoMove(ctx, game)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					fmt.Printf("Ошибка автохода: %v\n", err)
					break
				}
				recordMoveTime(game, mover, duration)
				fmt.Println(notation)
				render.DisplayBoard(game, 1)
			}
			continue
		}

		// Обычный ход
		game.Mu.Lock()
		move, err := render.ParseMove(input, game.CurrentPlayer, game.Board)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			game.Mu.Unlock()
			continue
		}

		if !move.IsValid(game.Board) {
			fmt.Println("Ошибка: неверные координаты хода")
			game.Mu.Unlock()
			continue
		}

		if err := render.ApplyMove(game.Board, move); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			game.Mu.Unlock()
			continue
		}

		game.MakeMove(move)
		duration := time.Since(startTime)
		setMoveTimeUnsafe(game, move.Player, duration)
		game.Mu.Unlock()
		repository.Store(move)

		fmt.Println()
		render.DisplayBoard(game, 1)
	}
}

func autoMove(ctx context.Context, game *model.Game) (time.Duration, string, *model.Player, error) {
	startTime := time.Now()
	delay := time.Duration(2+rand.Intn(3)) * time.Second
	timer := time.NewTimer(delay)
	select {
	case <-ctx.Done():
		timer.Stop()
		return 0, "", nil, ctx.Err()
	case <-timer.C:
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()
	board := game.Board
	player := game.CurrentPlayer

	for row := 0; row < board.Size; row++ {
		for col := 0; col < board.Size; col++ {
			piece := board.GetCell(row, col)
			if piece == "" || !render.IsPlayerPiece(piece, player) {
				continue
			}

			var targetRow int
			if player.IsWhite() {
				targetRow = row - 1
			} else {
				targetRow = row + 1
			}

			if !board.IsValidPosition(targetRow, col) {
				continue
			}

			targetPiece := board.GetCell(targetRow, col)
			if targetPiece != "" && render.IsPlayerPiece(targetPiece, player) {
				continue
			}

			move := model.NewMove(row, col, targetRow, col, player, piece)

			if err := render.ApplyMove(board, move); err != nil {
				continue
			}
			game.MakeMove(move)
			repository.Store(move)

			fromCol := string(rune('a' + col))
			fromRow := board.Size - row
			toCol := string(rune('a' + col))
			toRow := board.Size - targetRow
			duration := time.Since(startTime)
			notation := fmt.Sprintf("Автоход: %s%d-%s%d", fromCol, fromRow, toCol, toRow)
			return duration, notation, player, nil
		}
	}

	return 0, "", player, fmt.Errorf("нет доступных ходов для %s", player.GetDisplayName())
}

func recordMoveTime(game *model.Game, player *model.Player, duration time.Duration) {
	if game == nil || player == nil {
		return
	}
	game.Mu.Lock()
	defer game.Mu.Unlock()
	setMoveTimeUnsafe(game, player, duration)
}

func setMoveTimeUnsafe(game *model.Game, player *model.Player, duration time.Duration) {
	game.LastMoveTime = duration
	if player.IsWhite() {
		game.LastWhiteTime = duration
		return
	}
	game.LastBlackTime = duration
}

func displayAllBoards(games []*model.Game) {
	for i, game := range games {
		render.DisplayBoard(game, i+1)
	}
}

func gameSimulator(ctx context.Context, manager *GameManager, updateCh chan<- []*model.Game) {
	var wg sync.WaitGroup
	for _, game := range manager.GetGames() {
		wg.Add(1)
		go func(g *model.Game) {
			defer wg.Done()
			simulateGame(ctx, g, manager, updateCh)
		}(game)
	}
	wg.Wait()
}

func simulateGame(ctx context.Context, game *model.Game, manager *GameManager, updateCh chan<- []*model.Game) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		game.Mu.RLock()
		inProgress := game.IsInProgress()
		game.Mu.RUnlock()
		if !inProgress {
			manager.RemoveFinishedGames()
			return
		}

		duration, _, mover, err := autoMove(ctx, game)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			game.Mu.Lock()
			game.Finish()
			game.Mu.Unlock()
			manager.RemoveFinishedGames()
			return
		}
		recordMoveTime(game, mover, duration)
		sendGameSnapshot(manager, updateCh)
	}
}

func sendGameSnapshot(manager *GameManager, updateCh chan<- []*model.Game) {
	select {
	case updateCh <- manager.GetGames():
	default:
	}
}

func boardRenderer(ctx context.Context, manager *GameManager, updateCh <-chan []*model.Game) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var snapshot []*model.Game
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updateCh:
			snapshot = update
		case <-ticker.C:
			if snapshot == nil {
				snapshot = manager.GetGames()
			}
			displayAllBoards(snapshot)
		}
	}
}

func sliceLogger(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case change := <-repository.SliceChangeChan:
			fmt.Printf("[SLICE] %s %s at %s (%s)\n",
				change.SliceType,
				change.Operation,
				change.Timestamp.Format(time.RFC3339),
				change.Details,
			)
		}
	}
}
