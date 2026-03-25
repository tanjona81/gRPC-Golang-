package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	userv1 "github.com/tanjona81/gRPC-Golang-/gen/go"
	"github.com/tanjona81/gRPC-Golang-/internal/service"
	internalgrpc "github.com/tanjona81/gRPC-Golang-/internal/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func main() {
	//  Create a TCP listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Printf("PANIC RECOVERED: %v", p)
			return status.Errorf(codes.Internal, "an unexpected error occurred")
		}),
	}

	// Initialize our gRPC server with the service logic
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(opts...),
			internalgrpc.TimeoutInterceptor(5*time.Second),
			internalgrpc.RequestIDInterceptor,
			internalgrpc.LoggingInterceptor,
		),
	)
	userSvc := &service.UserService{}

	userv1.RegisterUserServiceServer(server, &internalgrpc.Server{Service: userSvc})

	// "scan" the server and automatically load all methods
	reflection.Register(server)

	// Run server in a goroutine so it doesn't block
	go func() {
		log.Printf("Server listening at %v", lis.Addr())
		if err := server.Serve(lis); err != nil {
			log.Fatalf("[CRITICAL] failed to serve: %v", err)
		}
	}()

	// Create a channel to listen for OS signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Wait here until a signal is received
	<-quit
	log.Println("Shutdown signal received. Shutting down gracefully...")

	// Stop accepting new requests immediately,
	// but give existing ones 5 seconds to finish.
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("Server exited cleanly")
	case <-time.After(5 * time.Second):
		log.Println("Shutdown timed out; forcing stop")
		server.Stop()
	}

	log.Println("Goodbye")
}
