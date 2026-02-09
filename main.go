package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/repository"
	"github.com/imyakin/go_hw/internal/service"
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
	games := startGames()
	if len(games) == 0 {
		return
	}

	manager := NewGameManager()
	for _, game := range games {
		placePieces(game.Board, game.Board.Size)
		game.Start()
		manager.AddGame(game)
		repository.Store(game)
		repository.Store(game.Board)
		repository.Store(game.WhitePlayer)
		repository.Store(game.BlackPlayer)
	}

	// Start entity generation ticker
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				service.GenerateEntities(repository.Store)
			}
		}
	}()

	// Slice change logger
	logDone := make(chan struct{})
	go sliceLogger(logDone)

	if manager.GetGameCount() == 1 {
		game := manager.GetGames()[0]
		displayBoard(game, 1)
		gameLoop(game)
	} else {
		updateCh := make(chan []*model.Game, 16)
		renderDone := make(chan struct{})
		simDone := make(chan struct{})

		go gameSimulator(manager, updateCh, simDone)
		go boardRenderer(manager, updateCh, renderDone)

		// Wait until all games are finished
		for manager.GetGameCount() > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		close(simDone)
		close(renderDone)
	}

	// Stop ticker and print stats
	ticker.Stop()
	done <- true
	close(logDone)
	repository.PrintStats()
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

func displayBoard(game *model.Game, index int) {
	game.Mu.RLock()
	defer game.Mu.RUnlock()
	board := game.Board
	size := board.Size
	rowNumberWidth := len(fmt.Sprintf("%d", size))

	fmt.Printf("\nИгра #%d\n", index)
	fmt.Printf("Время хода: %s %v | %s %v\n",
		game.WhitePlayer.GetDisplayName(),
		game.LastWhiteTime,
		game.BlackPlayer.GetDisplayName(),
		game.LastBlackTime,
	)

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
		fmt.Printf("\n%s, ваш ход (формат: e2-e4 или 'exit' для выхода или 'Автоход' или 'Сдался'): ", game.CurrentPlayer.GetDisplayName())

		var input string
		startTime := time.Now()
		fmt.Scan(&input)

		// 1. exit / quit
		if input == "exit" || input == "quit" {
			game.Finish()
			fmt.Println("Игра завершена!")
			fmt.Printf("Всего ходов: %d\n", game.GetMoveCount())
			break
		}

		// 2. Сдался
		if strings.EqualFold(input, "Сдался") {
			resigned := game.CurrentPlayer
			game.Resign()
			fmt.Printf("%s сдался! Победил %s!\n", resigned.GetDisplayName(), game.Winner.GetDisplayName())
			break
		}

		// 3. Автоход
		if strings.EqualFold(input, "Автоход") {
			fmt.Print("Сколько автоходов сделать: ")
			var count int
			_, err := fmt.Scan(&count)
			if err != nil || count <= 0 {
				fmt.Println("Ошибка: введите положительное число")
				continue
			}
			for i := 0; i < count; i++ {
				if !game.IsInProgress() {
					break
				}
				duration, notation, mover, err := autoMove(game)
				if err != nil {
					fmt.Printf("Ошибка автохода: %v\n", err)
					break
				}
				recordMoveTime(game, mover, duration)
				fmt.Println(notation)
				displayBoard(game, 1)
			}
			continue
		}

		// 4. Обычный ход
		game.Mu.Lock()
		move, err := parseMove(input, game.CurrentPlayer, game.Board)
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

		if err := applyMove(game.Board, move); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			game.Mu.Unlock()
			continue
		}

		// Add move to game history and switch player
		game.MakeMove(move)
		duration := time.Since(startTime)
		setMoveTimeUnsafe(game, move.Player, duration)
		game.Mu.Unlock()

		fmt.Println()
		displayBoard(game, 1)
	}
}

func isPlayerPiece(piece string, player *model.Player) bool {
	if player.IsWhite() {
		return piece == "♔" || piece == "♕" || piece == "♖" || piece == "♗" || piece == "♘" || piece == "♙"
	}
	return piece == "♚" || piece == "♛" || piece == "♜" || piece == "♝" || piece == "♞" || piece == "♟"
}

func autoMove(game *model.Game) (time.Duration, string, *model.Player, error) {
	startTime := time.Now()
	// Random delay 2-4 seconds
	delay := time.Duration(2+rand.Intn(3)) * time.Second
	time.Sleep(delay)

	game.Mu.Lock()
	defer game.Mu.Unlock()
	board := game.Board
	player := game.CurrentPlayer

	for row := 0; row < board.Size; row++ {
		for col := 0; col < board.Size; col++ {
			piece := board.GetCell(row, col)
			if piece == "" || !isPlayerPiece(piece, player) {
				continue
			}

			// Determine target row
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
			if targetPiece != "" && isPlayerPiece(targetPiece, player) {
				continue
			}

			move := model.NewMove(row, col, targetRow, col, player, piece)

			if err := applyMove(board, move); err != nil {
				continue
			}
			game.MakeMove(move)

			// Format move notation
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
		displayBoard(game, i+1)
	}
}

func gameSimulator(manager *GameManager, updateCh chan<- []*model.Game, done <-chan struct{}) {
	var wg sync.WaitGroup
	for _, game := range manager.GetGames() {
		wg.Add(1)
		go func(g *model.Game) {
			defer wg.Done()
			simulateGame(g, manager, updateCh, done)
		}(game)
	}
	wg.Wait()
}

func simulateGame(game *model.Game, manager *GameManager, updateCh chan<- []*model.Game, done <-chan struct{}) {
	for {
		select {
		case <-done:
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

		duration, _, mover, err := autoMove(game)
		if err != nil {
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

func boardRenderer(manager *GameManager, updateCh <-chan []*model.Game, done <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var snapshot []*model.Game
	for {
		select {
		case <-done:
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

func sliceLogger(done <-chan struct{}) {
	for {
		select {
		case <-done:
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
