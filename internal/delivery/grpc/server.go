package grpc

import (
	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	server     *grpc.Server
	taskServer pb.TaskServiceServer
}

func NewServer(taskService domain.TaskService) *Server {
	taskServer := NewTaskHandler(taskService)

	grpcServer := grpc.NewServer()

	pb.RegisterTaskServiceServer(grpcServer, taskServer)
	return &Server{
		server:     grpcServer,
		taskServer: taskServer,
	}
}

func (s *Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("grpc server listening on %s", address)

	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
