// Package grpcconf helps configure gRPC service connections.
package grpcconf

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Service defines a generic base for dealing with connection details
// to an internal gRPC service.
type Service struct {
	Host string `json:"host"`
	Port string `json:"port"`

	// Insecure is required by gRPC to say when there are no TLS connection
	// details.
	Insecure bool `json:"insecure"`

	// PublicURL defines the base url to use when forwarding the user to
	// public side of the service. Not all services required this.
	PublicURL string `json:"public_url"`
}

// DialOptions provides an array of gRPC DialOptions based on the
// defined service configuration.
func (s *Service) DialOptions() []grpc.DialOption {
	opts := []grpc.DialOption{}
	if s.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		))
	}
	return opts
}

// URL provides the result of joining the Host and Port together.
func (s *Service) URL() string {
	return s.Host + ":" + s.Port
}

// Connection provides an instance of the grpc connection.
func (s *Service) Connection(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	do := s.DialOptions()
	do = append(do, opts...)
	return grpc.NewClient(s.URL(), do...)
}
