// Package grpcconf helps configure gRPC service connections.
package grpcconf

import (
	"time"

	"google.golang.org/grpc"
	// Register the round_robin balancer so Policy == PolicyRoundRobin resolves
	// at runtime. The core grpc package already pulls this in transitively, but
	// importing it explicitly means an upstream change can't silently drop the
	// balancer and revert us to pick_first.
	_ "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Load-balancing policies accepted by Service.Policy.
const (
	// PolicyPickFirst opens a single connection to one backend and stays on it
	// for the life of the connection. This is gRPC's default and the zero value.
	PolicyPickFirst = "pick_first"

	// PolicyRoundRobin spreads calls across every backend address the resolver
	// returns and drops failed sub-channels on re-resolution. Use it when the
	// target is a *headless* Kubernetes Service (so DNS returns every pod IP);
	// it is what stops a single rolled pod from taking out the client.
	PolicyRoundRobin = "round_robin"
)

// roundRobinServiceConfig is the gRPC service config that selects the
// round_robin balancer (registered via the blank import above).
const roundRobinServiceConfig = `{"loadBalancingConfig":[{"round_robin":{}}]}`

const (
	// keepaliveTime is how long the connection can be idle before the client
	// sends a keepalive ping. It must stay >= the server's keepalive
	// EnforcementPolicy.MinTime (10s in our services): pinging faster earns
	// "ping strikes" and, after three, a too_many_pings GOAWAY that tears the
	// connection down. 20s leaves margin above that floor while detecting a
	// dead peer in well under a minute.
	//
	// REQUIRES the servers to run the relaxed EnforcementPolicy (MinTime 10s,
	// PermitWithoutStream). Deploy those first; against the gRPC default policy
	// (MinTime 5m) this interval would get connections GOAWAY'd.
	keepaliveTime = 20 * time.Second

	// keepaliveTimeout is how long to wait for a ping ack before considering
	// the connection dead and closing it. This is what bounds a request stuck
	// on a black-holed connection: without keepalive it hangs until the kernel
	// TCP timeout (~15m); with it, ~keepaliveTime+keepaliveTimeout.
	keepaliveTimeout = 10 * time.Second
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

	// Policy selects the client-side load-balancing policy: PolicyPickFirst
	// (the default, and the zero value) or PolicyRoundRobin. Any other value
	// falls back to gRPC's default (pick_first).
	Policy string `json:"policy"`
}

// DialOptions provides an array of gRPC DialOptions based on the
// defined service configuration.
func (s *Service) DialOptions() []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    keepaliveTime,
			Timeout: keepaliveTimeout,
			// Ping even with no active RPCs so idle pooled connections to a
			// rolled pod are detected and dropped before the next request lands
			// on them. Requires the servers' EnforcementPolicy to set
			// PermitWithoutStream too, else the server GOAWAYs on idle pings.
			PermitWithoutStream: true,
		}),
	}
	if s.Policy == PolicyRoundRobin {
		opts = append(opts, grpc.WithDefaultServiceConfig(roundRobinServiceConfig))
	}
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

// Connection provides an instance of the grpc connection. Caller-supplied opts
// are applied last, so a caller can override any of the defaults above.
func (s *Service) Connection(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	do := s.DialOptions()
	do = append(do, opts...)
	return grpc.NewClient(s.URL(), do...)
}
