package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// User is our internal domain model
type User struct {
	ID       string
	Email    string
	FullName string
}

// UserService defines the business logic contract
type UserService struct{}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
	// For now, we mock the database logic
	if id == "" || id == "0" {
		return nil, status.Error(codes.InvalidArgument, "user id cannot be empty or zero")
	}

	// Mocking a "Not Found" scenario
	if id == "999" {
		return nil, status.Error(codes.NotFound, "user not found in database")
	}
	return &User{
		ID:       id,
		Email:    "test@test.gmail",
		FullName: "Jhon Test",
	}, nil
}
