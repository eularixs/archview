package main

import (
	"archview-example-microservices/server"
	"archview-example-microservices/spec"
)

func main() {
	var reg spec.ServiceRegistrar
	spec.RegisterUserServiceServer(reg, server.UserServer{})
}
