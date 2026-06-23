package main

import (
	"context"

	"archview-example-microservices/spec"
)

func main() {
	var cc *spec.ClientConn
	client := spec.NewUserServiceClient(cc)
	_, _ = client.GetUser(context.Background(), &spec.GetUserRequest{Id: "1"})
}
