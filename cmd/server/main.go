package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imyakin/go_hw/internal/handler"
	"github.com/imyakin/go_hw/internal/repository"
)

const boardPageTemplate = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="refresh" content="3">
    <title>Шахматные доски — Наблюдатель</title>
    <style>
        body { background: #1a1a2e; color: #eee; font-family: monospace; padding: 20px; }
        h1 { color: #e94560; }
        .game { background: #16213e; border: 1px solid #0f3460; border-radius: 8px; padding: 16px; margin: 16px 0; display: inline-block; margin-right: 16px; vertical-align: top; }
        .game h2 { color: #e94560; margin-top: 0; }
        .info { color: #a8a8a8; margin-bottom: 8px; }
        .status { font-weight: bold; }
        .status-in_progress { color: #4ecca3; }
        .status-finished { color: #e94560; }
        .status-not_started { color: #a8a8a8; }
        table { border-collapse: collapse; }
        td { width: 32px; height: 32px; text-align: center; font-size: 20px; }
        .light { background: #f0d9b5; color: #000; }
        .dark { background: #b58863; color: #000; }
        .col-header, .row-header { background: transparent; color: #a8a8a8; font-size: 14px; border: none; }
    </style>
</head>
<body>
    <h1>Шахматные доски</h1>
    <p>Страница обновляется автоматически каждые 3 секунды</p>
    {{if eq (len .Games) 0}}
        <p>Нет активных игр. Создайте игру через POST /api/games</p>
    {{end}}
    {{range .Games}}
    <div class="game">
        <h2>Игра #{{.ID}}</h2>
        <div class="info">
            ♔ {{.WhitePlayerName}} vs ♚ {{.BlackPlayerName}}
        </div>
        <div class="info">
            Статус: <span class="status status-{{.Status}}">{{.Status}}</span>
            {{if .CurrentPlayerColor}}| Ход: {{.CurrentPlayerColor}}{{end}}
            {{if .WinnerColor}}| Победитель: {{.WinnerColor}}{{end}}
        </div>
        <table>
            <tr>
                <td class="col-header"></td>
                {{range .ColHeaders}}<td class="col-header">{{.}}</td>{{end}}
            </tr>
            {{range .Rows}}
            <tr>
                <td class="row-header">{{.RowNum}}</td>
                {{range .Cells}}
                <td class="{{.Class}}">{{.Piece}}</td>
                {{end}}
            </tr>
            {{end}}
        </table>
    </div>
    {{end}}
</body>
</html>`

type boardPageData struct {
	Games []gamePageData
}

type gamePageData struct {
	ID                 int
	WhitePlayerName    string
	BlackPlayerName    string
	Status             string
	CurrentPlayerColor string
	WinnerColor        string
	ColHeaders         []string
	Rows               []rowData
}

type rowData struct {
	RowNum int
	Cells  []cellData
}

type cellData struct {
	Piece string
	Class string
}

func main() {
	if err := repository.LoadAll(); err != nil {
		fmt.Printf("Предупреждение: ошибка загрузки данных из CSV: %v\n", err)
	} else {
		fmt.Println("Данные загружены из CSV файлов.")
		repository.PrintStats()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go sliceLogger(ctx)

	r := gin.Default()

	// HTML page for observer
	tmpl := template.Must(template.New("boards").Parse(boardPageTemplate))
	r.GET("/", func(c *gin.Context) {
		games := repository.GetGames()
		data := boardPageData{Games: make([]gamePageData, 0, len(games))}

		for _, g := range games {
			g.Mu.RLock()
			gpd := gamePageData{
				ID:              g.ID,
				WhitePlayerName: g.WhitePlayer.Name,
				BlackPlayerName: g.BlackPlayer.Name,
				Status:          string(g.Status),
			}
			if g.CurrentPlayer != nil {
				gpd.CurrentPlayerColor = string(g.CurrentPlayer.Color)
			}
			if g.Winner != nil {
				gpd.WinnerColor = string(g.Winner.Color)
			}

			size := g.Board.Size
			for j := 0; j < size; j++ {
				gpd.ColHeaders = append(gpd.ColHeaders, string(rune('A'+j%26)))
			}

			for i := 0; i < size; i++ {
				row := rowData{RowNum: size - i}
				for j := 0; j < size; j++ {
					piece := g.Board.GetCell(i, j)
					class := "light"
					if (i+j)%2 != 0 {
						class = "dark"
					}
					row.Cells = append(row.Cells, cellData{Piece: piece, Class: class})
				}
				gpd.Rows = append(gpd.Rows, row)
			}
			g.Mu.RUnlock()
			data.Games = append(data.Games, gpd)
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(c.Writer, data)
	})

	// JSON API
	api := r.Group("/api")

	api.POST("/games", handler.CreateGame)
	api.GET("/games", handler.ListGames)
	api.GET("/games/:id", handler.GetGame)
	api.PUT("/games/:id", handler.UpdateGame)

	api.POST("/boards", handler.CreateBoard)
	api.GET("/boards", handler.ListBoards)
	api.GET("/boards/:id", handler.GetBoard)
	api.PUT("/boards/:id", handler.UpdateBoard)

	api.POST("/players", handler.CreatePlayer)
	api.GET("/players", handler.ListPlayers)
	api.GET("/players/:id", handler.GetPlayer)
	api.PUT("/players/:id", handler.UpdatePlayer)

	api.POST("/moves", handler.CreateMove)
	api.GET("/moves", handler.ListMoves)
	api.GET("/moves/:id", handler.GetMove)
	api.PUT("/moves/:id", handler.UpdateMove)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		fmt.Println("Сервер запущен на http://localhost:8080")
		fmt.Println("Наблюдатель: http://localhost:8080/")
		fmt.Println("API: http://localhost:8080/api/")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка сервера: %v\n", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nЗавершение работы сервера...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Ошибка при graceful shutdown: %v\n", err)
	}

	if err := repository.SaveAll(); err != nil {
		fmt.Printf("Ошибка сохранения данных: %v\n", err)
	}

	fmt.Println("Сервер остановлен.")
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

