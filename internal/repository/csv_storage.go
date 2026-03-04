package repository

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/imyakin/go_hw/internal/model"
)

const dataDir = "data"

func ensureDataDir() {
	os.MkdirAll(dataDir, 0755)
}

func savePlayersCSV() error {
	ensureDataDir()
	f, err := os.Create(filepath.Join(dataDir, "players.csv"))
	if err != nil {
		return fmt.Errorf("create players.csv: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"Name", "Color", "Symbol"}); err != nil {
		return err
	}
	for _, p := range players {
		if err := w.Write([]string{p.Name, string(p.Color), p.Symbol}); err != nil {
			return err
		}
	}
	return w.Error()
}

func loadPlayers() ([]*model.Player, error) {
	f, err := os.Open(filepath.Join(dataDir, "players.csv"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []*model.Player
	for i, rec := range records {
		// i == 0 — это строка заголовка (Name, Color, Symbol), её пропускаем.
		// Также пропускаем строки, в которых меньше 3 полей, чтобы не обращаться к несуществующим индексам.
		if i == 0 {
			continue
		}
		if len(rec) < 3 {
			continue
		}
		result = append(result, &model.Player{
			Name:   rec[0],
			Color:  model.PlayerColor(rec[1]),
			Symbol: rec[2],
		})
	}
	return result, nil
}

func saveBoardsCSV() error {
	ensureDataDir()
	f, err := os.Create(filepath.Join(dataDir, "boards.csv"))
	if err != nil {
		return fmt.Errorf("create boards.csv: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"Size", "Cells"}); err != nil {
		return err
	}
	for _, b := range boards {
		cellsJSON, err := json.Marshal(b.Cells)
		if err != nil {
			return fmt.Errorf("marshal board cells: %w", err)
		}
		if err := w.Write([]string{strconv.Itoa(b.Size), string(cellsJSON)}); err != nil {
			return err
		}
	}
	return w.Error()
}

func loadBoards() ([]*model.Board, error) {
	f, err := os.Open(filepath.Join(dataDir, "boards.csv"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []*model.Board
	for i, rec := range records {
		// i == 0 — строка заголовка (Size, Cells), её пропускаем.
		// Также пропускаем строки, в которых меньше 2 полей.
		if i == 0 {
			continue
		}
		if len(rec) < 2 {
			continue
		}
		size, err := strconv.Atoi(rec[0])
		if err != nil {
			continue
		}
		var cells [][]string
		if err := json.Unmarshal([]byte(rec[1]), &cells); err != nil {
			continue
		}
		result = append(result, &model.Board{
			Size:  size,
			Cells: cells,
		})
	}
	return result, nil
}

func saveMovesCSV() error {
	ensureDataDir()
	f, err := os.Create(filepath.Join(dataDir, "moves.csv"))
	if err != nil {
		return fmt.Errorf("create moves.csv: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"FromRow", "FromCol", "ToRow", "ToCol", "PlayerName", "PlayerColor", "Piece"}); err != nil {
		return err
	}
	for _, m := range moves {
		playerName, playerColor := "", ""
		if m.Player != nil {
			playerName = m.Player.Name
			playerColor = string(m.Player.Color)
		}
		if err := w.Write([]string{
			strconv.Itoa(m.From.Row),
			strconv.Itoa(m.From.Col),
			strconv.Itoa(m.To.Row),
			strconv.Itoa(m.To.Col),
			playerName,
			playerColor,
			m.Piece,
		}); err != nil {
			return err
		}
	}
	return w.Error()
}

func loadMoves() ([]*model.Move, error) {
	f, err := os.Open(filepath.Join(dataDir, "moves.csv"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []*model.Move
	for i, rec := range records {
		if i == 0 {
			continue
		}
		if len(rec) < 7 {
			continue
		}
		fromRow, _ := strconv.Atoi(rec[0])
		fromCol, _ := strconv.Atoi(rec[1])
		toRow, _ := strconv.Atoi(rec[2])
		toCol, _ := strconv.Atoi(rec[3])

		var player *model.Player
		if rec[4] != "" {
			player = model.NewPlayer(rec[4], model.PlayerColor(rec[5]))
		}

		result = append(result, &model.Move{
			From:   model.Position{Row: fromRow, Col: fromCol},
			To:     model.Position{Row: toRow, Col: toCol},
			Player: player,
			Piece:  rec[6],
		})
	}
	return result, nil
}

func saveGamesCSV() error {
	ensureDataDir()
	f, err := os.Create(filepath.Join(dataDir, "games.csv"))
	if err != nil {
		return fmt.Errorf("create games.csv: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{
		"WhitePlayerName", "BlackPlayerName", "BoardSize",
		"Status", "CurrentPlayerColor", "WinnerColor", "Cells",
	}); err != nil {
		return err
	}

	for _, g := range games {
		g.Mu.RLock()
		currentColor := ""
		if g.CurrentPlayer != nil {
			currentColor = string(g.CurrentPlayer.Color)
		}
		winnerColor := ""
		if g.Winner != nil {
			winnerColor = string(g.Winner.Color)
		}
		cellsJSON, _ := json.Marshal(g.Board.Cells)
		err := w.Write([]string{
			g.WhitePlayer.Name,
			g.BlackPlayer.Name,
			strconv.Itoa(g.Board.Size),
			string(g.Status),
			currentColor,
			winnerColor,
			string(cellsJSON),
		})
		g.Mu.RUnlock()
		if err != nil {
			return err
		}
	}
	return w.Error()
}

func loadGames() ([]*model.Game, error) {
	f, err := os.Open(filepath.Join(dataDir, "games.csv"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []*model.Game
	for i, rec := range records {
		if i == 0 {
			continue
		}
		if len(rec) < 7 {
			continue
		}

		boardSize, _ := strconv.Atoi(rec[2])
		var cells [][]string
		json.Unmarshal([]byte(rec[6]), &cells)

		whitePlayer := model.NewPlayer(rec[0], model.White)
		blackPlayer := model.NewPlayer(rec[1], model.Black)

		board := &model.Board{Size: boardSize, Cells: cells}

		game := &model.Game{
			WhitePlayer: whitePlayer,
			BlackPlayer: blackPlayer,
			Board:       board,
			Moves:       make([]*model.Move, 0),
			Status:      model.GameStatus(rec[3]),
		}

		if rec[4] == string(model.Black) {
			game.CurrentPlayer = blackPlayer
		} else {
			game.CurrentPlayer = whitePlayer
		}

		if rec[5] == string(model.White) {
			game.Winner = whitePlayer
		} else if rec[5] == string(model.Black) {
			game.Winner = blackPlayer
		}

		result = append(result, game)
	}
	return result, nil
}

func SaveAll() error {
	var errs []error

	muPlayers.RLock()
	if err := savePlayersCSV(); err != nil {
		errs = append(errs, err)
	}
	muPlayers.RUnlock()

	muBoards.RLock()
	if err := saveBoardsCSV(); err != nil {
		errs = append(errs, err)
	}
	muBoards.RUnlock()

	muMoves.RLock()
	if err := saveMovesCSV(); err != nil {
		errs = append(errs, err)
	}
	muMoves.RUnlock()

	muGames.RLock()
	if err := saveGamesCSV(); err != nil {
		errs = append(errs, err)
	}
	muGames.RUnlock()

	return errors.Join(errs...)
}

func LoadAll() error {
	var errs []error

	loaded, err := loadPlayers()
	if err != nil {
		errs = append(errs, fmt.Errorf("load players: %w", err))
	} else if loaded != nil {
		muPlayers.Lock()
		players = loaded
		muPlayers.Unlock()
		notifySliceChange("players", "load", fmt.Sprintf("loaded %d players from CSV", len(loaded)))
	}

	loadedBoards, err := loadBoards()
	if err != nil {
		errs = append(errs, fmt.Errorf("load boards: %w", err))
	} else if loadedBoards != nil {
		muBoards.Lock()
		boards = loadedBoards
		muBoards.Unlock()
		notifySliceChange("boards", "load", fmt.Sprintf("loaded %d boards from CSV", len(loadedBoards)))
	}

	loadedMoves, err := loadMoves()
	if err != nil {
		errs = append(errs, fmt.Errorf("load moves: %w", err))
	} else if loadedMoves != nil {
		muMoves.Lock()
		moves = loadedMoves
		muMoves.Unlock()
		notifySliceChange("moves", "load", fmt.Sprintf("loaded %d moves from CSV", len(loadedMoves)))
	}

	loadedGames, err := loadGames()
	if err != nil {
		errs = append(errs, fmt.Errorf("load games: %w", err))
	} else if loadedGames != nil {
		muGames.Lock()
		games = loadedGames
		muGames.Unlock()
		notifySliceChange("games", "load", fmt.Sprintf("loaded %d games from CSV", len(loadedGames)))
	}

	return errors.Join(errs...)
}
