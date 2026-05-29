package grpcserver

import (
	"context"

	"github.com/imyakin/go_hw/internal/model"
	"github.com/imyakin/go_hw/internal/render"
	"github.com/imyakin/go_hw/internal/repository"
	chessv1 "github.com/imyakin/go_hw/proto/chess/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- cells <-> rows ---

func rowsToCells(rows []*chessv1.Row) [][]string {
	if len(rows) == 0 {
		return nil
	}
	cells := make([][]string, len(rows))
	for i, r := range rows {
		if r == nil {
			continue
		}
		cells[i] = append([]string(nil), r.Cells...)
	}
	return cells
}

func cellsToRows(cells [][]string) []*chessv1.Row {
	if len(cells) == 0 {
		return nil
	}
	rows := make([]*chessv1.Row, len(cells))
	for i, row := range cells {
		rows[i] = &chessv1.Row{Cells: append([]string(nil), row...)}
	}
	return rows
}

func playerColorToProto(c model.PlayerColor) chessv1.PlayerColor {
	switch c {
	case model.White:
		return chessv1.PlayerColor_PLAYER_COLOR_WHITE
	case model.Black:
		return chessv1.PlayerColor_PLAYER_COLOR_BLACK
	default:
		return chessv1.PlayerColor_PLAYER_COLOR_UNSPECIFIED
	}
}

func playerColorFromProto(c chessv1.PlayerColor) model.PlayerColor {
	switch c {
	case chessv1.PlayerColor_PLAYER_COLOR_BLACK:
		return model.Black
	default:
		return model.White
	}
}

func gameStatusToProto(s model.GameStatus) chessv1.GameStatus {
	switch s {
	case model.StatusNotStarted:
		return chessv1.GameStatus_GAME_STATUS_NOT_STARTED
	case model.StatusInProgress:
		return chessv1.GameStatus_GAME_STATUS_IN_PROGRESS
	case model.StatusFinished:
		return chessv1.GameStatus_GAME_STATUS_FINISHED
	default:
		return chessv1.GameStatus_GAME_STATUS_UNSPECIFIED
	}
}

func gameStatusFromProto(s chessv1.GameStatus) model.GameStatus {
	switch s {
	case chessv1.GameStatus_GAME_STATUS_IN_PROGRESS:
		return model.StatusInProgress
	case chessv1.GameStatus_GAME_STATUS_FINISHED:
		return model.StatusFinished
	default:
		return model.StatusNotStarted
	}
}

func playerToProto(p *model.Player) *chessv1.Player {
	if p == nil {
		return nil
	}
	return &chessv1.Player{
		Id:     int32(p.ID),
		Name:   p.Name,
		Color:  playerColorToProto(p.Color),
		Symbol: p.Symbol,
	}
}

func boardToProto(b *model.Board) *chessv1.Board {
	if b == nil {
		return nil
	}
	return &chessv1.Board{
		Id:   int32(b.ID),
		Size: int32(b.Size),
		Rows: cellsToRows(b.Cells),
	}
}

func gameToProto(g *model.Game) *chessv1.Game {
	g.Mu.RLock()
	defer g.Mu.RUnlock()

	current := chessv1.PlayerColor_PLAYER_COLOR_UNSPECIFIED
	if g.CurrentPlayer != nil {
		current = playerColorToProto(g.CurrentPlayer.Color)
	}
	winner := chessv1.PlayerColor_PLAYER_COLOR_UNSPECIFIED
	if g.Winner != nil {
		winner = playerColorToProto(g.Winner.Color)
	}

	return &chessv1.Game{
		Id:                 int32(g.ID),
		WhitePlayerName:    g.WhitePlayer.Name,
		BlackPlayerName:    g.BlackPlayer.Name,
		BoardSize:          int32(g.Board.Size),
		Status:             gameStatusToProto(g.Status),
		CurrentPlayerColor: current,
		WinnerColor:        winner,
		Rows:               cellsToRows(g.Board.Cells),
	}
}

func moveToProto(m *model.Move) *chessv1.Move {
	playerName := ""
	if m.Player != nil {
		playerName = m.Player.GetDisplayName()
	}
	return &chessv1.Move{
		Id:       int32(m.ID),
		GameId:   int32(m.GameID),
		FromRow:  int32(m.From.Row),
		FromCol:  int32(m.From.Col),
		ToRow:    int32(m.To.Row),
		ToCol:    int32(m.To.Col),
		Piece:    m.Piece,
		Notation: m.GetNotation(),
		Player:   playerName,
	}
}

// --- PlayerService ---

type PlayerServer struct {
	chessv1.UnimplementedPlayerServiceServer
}

func (s *PlayerServer) CreatePlayer(_ context.Context, req *chessv1.CreatePlayerRequest) (*chessv1.Player, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	p := model.NewPlayer(req.GetName(), playerColorFromProto(req.GetColor()))
	repository.Store(p)
	return playerToProto(p), nil
}

func (s *PlayerServer) GetPlayer(_ context.Context, req *chessv1.GetByIdRequest) (*chessv1.Player, error) {
	p := repository.GetPlayerByID(int(req.GetId()))
	if p == nil {
		return nil, status.Error(codes.NotFound, "player not found")
	}
	return playerToProto(p), nil
}

func (s *PlayerServer) ListPlayers(_ context.Context, _ *chessv1.ListRequest) (*chessv1.ListPlayersResponse, error) {
	players := repository.GetPlayers()
	items := make([]*chessv1.Player, 0, len(players))
	for _, p := range players {
		items = append(items, playerToProto(p))
	}
	return &chessv1.ListPlayersResponse{Items: items}, nil
}

func (s *PlayerServer) UpdatePlayer(_ context.Context, req *chessv1.UpdatePlayerRequest) (*chessv1.Player, error) {
	id := int(req.GetId())
	found := repository.UpdatePlayer(id, func(p *model.Player) {
		p.Name = req.GetName()
		p.Color = playerColorFromProto(req.GetColor())
		if p.Color == model.Black {
			p.Symbol = "♚"
		} else {
			p.Symbol = "♔"
		}
	})
	if !found {
		return nil, status.Error(codes.NotFound, "player not found")
	}
	return playerToProto(repository.GetPlayerByID(id)), nil
}

func (s *PlayerServer) DeletePlayer(context.Context, *chessv1.DeleteByIdRequest) (*chessv1.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "delete not implemented yet")
}

// --- BoardService ---

type BoardServer struct {
	chessv1.UnimplementedBoardServiceServer
}

func (s *BoardServer) CreateBoard(_ context.Context, req *chessv1.CreateBoardRequest) (*chessv1.Board, error) {
	if req.GetSize() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "size must be > 0")
	}
	b := model.NewBoard(int(req.GetSize()))
	repository.Store(b)
	return boardToProto(b), nil
}

func (s *BoardServer) GetBoard(_ context.Context, req *chessv1.GetByIdRequest) (*chessv1.Board, error) {
	b := repository.GetBoardByID(int(req.GetId()))
	if b == nil {
		return nil, status.Error(codes.NotFound, "board not found")
	}
	return boardToProto(b), nil
}

func (s *BoardServer) ListBoards(_ context.Context, _ *chessv1.ListRequest) (*chessv1.ListBoardsResponse, error) {
	boards := repository.GetBoards()
	items := make([]*chessv1.Board, 0, len(boards))
	for _, b := range boards {
		items = append(items, boardToProto(b))
	}
	return &chessv1.ListBoardsResponse{Items: items}, nil
}

func (s *BoardServer) UpdateBoard(_ context.Context, req *chessv1.UpdateBoardRequest) (*chessv1.Board, error) {
	id := int(req.GetId())
	found := repository.UpdateBoard(id, func(b *model.Board) {
		b.Size = int(req.GetSize())
		b.Cells = rowsToCells(req.GetRows())
	})
	if !found {
		return nil, status.Error(codes.NotFound, "board not found")
	}
	return boardToProto(repository.GetBoardByID(id)), nil
}

func (s *BoardServer) DeleteBoard(context.Context, *chessv1.DeleteByIdRequest) (*chessv1.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "delete not implemented yet")
}

// --- GameService ---

type GameServer struct {
	chessv1.UnimplementedGameServiceServer
}

func (s *GameServer) CreateGame(_ context.Context, req *chessv1.CreateGameRequest) (*chessv1.Game, error) {
	if req.GetWhitePlayerName() == "" || req.GetBlackPlayerName() == "" {
		return nil, status.Error(codes.InvalidArgument, "player names are required")
	}
	if req.GetBoardSize() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "board_size must be > 0")
	}

	game := model.NewGame(req.GetWhitePlayerName(), req.GetBlackPlayerName(), int(req.GetBoardSize()))
	render.PlacePieces(game.Board, int(req.GetBoardSize()))
	game.Start()

	repository.Store(game)
	repository.Store(game.Board)
	repository.Store(game.WhitePlayer)
	repository.Store(game.BlackPlayer)

	return gameToProto(game), nil
}

func (s *GameServer) GetGame(_ context.Context, req *chessv1.GetByIdRequest) (*chessv1.Game, error) {
	g := repository.GetGameByID(int(req.GetId()))
	if g == nil {
		return nil, status.Error(codes.NotFound, "game not found")
	}
	return gameToProto(g), nil
}

func (s *GameServer) ListGames(_ context.Context, _ *chessv1.ListRequest) (*chessv1.ListGamesResponse, error) {
	games := repository.GetGames()
	items := make([]*chessv1.Game, 0, len(games))
	for _, g := range games {
		items = append(items, gameToProto(g))
	}
	return &chessv1.ListGamesResponse{Items: items}, nil
}

func (s *GameServer) UpdateGame(_ context.Context, req *chessv1.UpdateGameRequest) (*chessv1.Game, error) {
	id := int(req.GetId())
	found := repository.UpdateGame(id, func(g *model.Game) {
		g.WhitePlayer.Name = req.GetWhitePlayerName()
		g.BlackPlayer.Name = req.GetBlackPlayerName()
		g.Status = gameStatusFromProto(req.GetStatus())
	})
	if !found {
		return nil, status.Error(codes.NotFound, "game not found")
	}
	return gameToProto(repository.GetGameByID(id)), nil
}

func (s *GameServer) DeleteGame(context.Context, *chessv1.DeleteByIdRequest) (*chessv1.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "delete not implemented yet")
}

// --- MoveService ---

type MoveServer struct {
	chessv1.UnimplementedMoveServiceServer
}

func (s *MoveServer) CreateMove(_ context.Context, req *chessv1.CreateMoveRequest) (*chessv1.Move, error) {
	game := repository.GetGameByID(int(req.GetGameId()))
	if game == nil {
		return nil, status.Error(codes.NotFound, "game not found")
	}

	game.Mu.Lock()

	if !game.IsInProgress() {
		game.Mu.Unlock()
		return nil, status.Error(codes.FailedPrecondition, "game is not in progress")
	}

	var move *model.Move
	if req.GetNotation() != "" {
		var err error
		move, err = render.ParseMove(req.GetNotation(), game.CurrentPlayer, game.Board)
		if err != nil {
			game.Mu.Unlock()
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		piece := game.Board.GetCell(int(req.GetFromRow()), int(req.GetFromCol()))
		if piece == "" {
			game.Mu.Unlock()
			return nil, status.Error(codes.InvalidArgument, "no piece at from position")
		}
		move = model.NewMove(
			int(req.GetFromRow()), int(req.GetFromCol()),
			int(req.GetToRow()), int(req.GetToCol()),
			game.CurrentPlayer, piece,
		)
	}

	if !move.IsValid(game.Board) {
		game.Mu.Unlock()
		return nil, status.Error(codes.InvalidArgument, "invalid move coordinates")
	}
	if err := render.ApplyMove(game.Board, move); err != nil {
		game.Mu.Unlock()
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	game.MakeMove(move)
	move.GameID = int(req.GetGameId())
	gameID := game.ID
	game.Mu.Unlock()

	repository.Store(move)
	repository.UpdateGame(gameID, func(g *model.Game) {})

	return moveToProto(move), nil
}

func (s *MoveServer) GetMove(_ context.Context, req *chessv1.GetByIdRequest) (*chessv1.Move, error) {
	m := repository.GetMoveByID(int(req.GetId()))
	if m == nil {
		return nil, status.Error(codes.NotFound, "move not found")
	}
	return moveToProto(m), nil
}

func (s *MoveServer) ListMoves(_ context.Context, req *chessv1.ListMovesRequest) (*chessv1.ListMovesResponse, error) {
	all := repository.GetMoves()
	items := make([]*chessv1.Move, 0)
	for _, m := range all {
		if req.GetGameId() != 0 && m.GameID != int(req.GetGameId()) {
			continue
		}
		items = append(items, moveToProto(m))
	}
	return &chessv1.ListMovesResponse{Items: items}, nil
}

func (s *MoveServer) UpdateMove(_ context.Context, req *chessv1.UpdateMoveRequest) (*chessv1.Move, error) {
	id := int(req.GetId())
	found := repository.UpdateMove(id, func(m *model.Move) {
		m.GameID = int(req.GetGameId())
		m.From = model.Position{Row: int(req.GetFromRow()), Col: int(req.GetFromCol())}
		m.To = model.Position{Row: int(req.GetToRow()), Col: int(req.GetToCol())}
		m.Piece = req.GetPiece()
	})
	if !found {
		return nil, status.Error(codes.NotFound, "move not found")
	}
	return moveToProto(repository.GetMoveByID(id)), nil
}

func (s *MoveServer) DeleteMove(context.Context, *chessv1.DeleteByIdRequest) (*chessv1.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "delete not implemented yet")
}
