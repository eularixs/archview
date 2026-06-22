// Package grpcrt is a tiny stand-in for google.golang.org/grpc so the example
// stays dependency-free. archview detects gRPC structurally (by the
// Register<Svc>Server shape), so the same detection works against real grpc.
package grpcrt

// ServiceRegistrar registers a service implementation (mirrors grpc.ServiceRegistrar).
type ServiceRegistrar interface {
	RegisterService(desc *ServiceDesc, impl any)
}

// ServiceDesc describes a service (mirrors grpc.ServiceDesc).
type ServiceDesc struct {
	ServiceName string
	HandlerType any
}

// Server is a minimal registrar (mirrors *grpc.Server).
type Server struct{ services []string }

// NewServer returns a Server.
func NewServer() *Server { return &Server{} }

// RegisterService records a registered service.
func (s *Server) RegisterService(desc *ServiceDesc, impl any) {
	s.services = append(s.services, desc.ServiceName)
}
