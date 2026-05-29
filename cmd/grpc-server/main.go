package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/imyakin/go_hw/internal/grpcserver"
	"github.com/imyakin/go_hw/internal/repository"
	chessv1 "github.com/imyakin/go_hw/proto/chess/v1"
	"google.golang.org/grpc"
)

func main() {
	if err := repository.LoadAll(); err != nil {
		fmt.Printf("Предупреждение: ошибка загрузки CSV: %v\n", err)
	} else {
		fmt.Println("Данные загружены из CSV.")
		repository.PrintStats()
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	chessv1.RegisterPlayerServiceServer(s, &grpcserver.PlayerServer{})
	chessv1.RegisterBoardServiceServer(s, &grpcserver.BoardServer{})
	chessv1.RegisterGameServiceServer(s, &grpcserver.GameServer{})
	chessv1.RegisterMoveServiceServer(s, &grpcserver.MoveServer{})

	go func() {
		fmt.Println("gRPC сервер слушает :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nОстановка gRPC сервера...")
	s.GracefulStop()

	if err := repository.SaveAll(); err != nil {
		fmt.Printf("Ошибка сохранения: %v\n", err)
	}
}
