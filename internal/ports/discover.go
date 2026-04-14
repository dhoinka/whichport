package ports

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
	ProtocolAll Protocol = "all"
)

type Query struct {
	Port      int
	Protocols []Protocol
}

type Listener struct {
	Port        int      `json:"port"`
	Protocol    Protocol `json:"protocol"`
	PID         int      `json:"pid"`
	Command     string   `json:"command"`
	CommandLine string   `json:"commandLine"`
	Path        string   `json:"path"`
}

func NewQuery(port int, protocol string) (Query, error) {
	if port < 0 || port > 65535 {
		return Query{}, fmt.Errorf("port must be between 0 and 65535")
	}

	switch normalized := strings.ToLower(strings.TrimSpace(protocol)); normalized {
	case "", string(ProtocolAll):
		return Query{Port: port, Protocols: []Protocol{ProtocolTCP, ProtocolUDP}}, nil
	case string(ProtocolTCP):
		return Query{Port: port, Protocols: []Protocol{ProtocolTCP}}, nil
	case string(ProtocolUDP):
		return Query{Port: port, Protocols: []Protocol{ProtocolUDP}}, nil
	default:
		return Query{}, fmt.Errorf("unsupported protocol %q: expected tcp, udp, or all", protocol)
	}
}

func Discover(ctx context.Context, query Query) ([]Listener, error) {
	listeners, err := discoverPlatform(ctx, query)
	if err != nil {
		return nil, err
	}

	sort.Slice(listeners, func(i, j int) bool {
		left := listeners[i]
		right := listeners[j]

		switch {
		case left.Port != right.Port:
			return left.Port < right.Port
		case left.Protocol != right.Protocol:
			return left.Protocol < right.Protocol
		case left.PID != right.PID:
			return left.PID < right.PID
		default:
			return left.Command < right.Command
		}
	})

	return listeners, nil
}
