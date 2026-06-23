package server

import (
	"context"

	"archview-example-microservices/spec"
)

type UserServer struct{}

func (UserServer) GetUser(ctx context.Context, r *spec.GetUserRequest) (*spec.GetUserResponse, error) {
	return &spec.GetUserResponse{Name: "u-" + r.Id}, nil
}
