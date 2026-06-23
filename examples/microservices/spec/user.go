// Package spec mimics protoc-gen-go-grpc output: a server-registration function
// and a typed client, the two shapes archview's gRPC stitching keys on.
package spec

import "context"

type GetUserRequest struct{ Id string }
type GetUserResponse struct{ Name string }

// Server side.
type UserServiceServer interface {
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
}

type ServiceRegistrar interface{ register() }

func RegisterUserServiceServer(s ServiceRegistrar, srv UserServiceServer) { _ = s; _ = srv }

// Client side.
type UserServiceClient interface {
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
}

type ClientConn struct{}

func NewUserServiceClient(cc *ClientConn) UserServiceClient { _ = cc; return nil }
