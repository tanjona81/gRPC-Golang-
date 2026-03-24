package grpc

import (
	"context"
	"log"

	userv1 "github.com/tanjona81/go-grpc/gen/go"
	"github.com/tanjona81/go-grpc/internal/service"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	Service *service.UserService
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	log.Printf("Received GetUser request for ID: %s", req.GetUserId())
	res, err := s.Service.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserResponse{
		Id:       res.ID,
		Email:    res.Email,
		FullName: res.FullName,
	}, nil
}
