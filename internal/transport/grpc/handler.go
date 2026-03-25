package grpc

import (
	"context"
	"log"
	"time"

	userv1 "github.com/tanjona81/gRPC-Golang-/gen/go"
	"github.com/tanjona81/gRPC-Golang-/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	Service *service.UserService
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	log.Printf("Received GetUser request for ID: %s", req.GetUserId())

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := s.Service.GetUser(ctx, req.GetUserId())

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, status.Error(codes.DeadlineExceeded, "database took too long")
		}
		return nil, status.Error(codes.Internal, "failed to fetch user")
	}

	return &userv1.GetUserResponse{
		Id:       res.ID,
		Email:    res.Email,
		FullName: res.FullName,
	}, nil
}
