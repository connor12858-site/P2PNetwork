// Package testapp provides a minimal request/reply app used to verify app streams.
package testapp

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"fgov/network/pkg/node"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ID       = "test"
	Name     = "Test App"
	Protocol = protocol.ID("/p2pnetwork/test/1.0.0")
)

// Register exposes the test app on a node.
func Register(n *node.Node) error {
	return n.RegisterApp(node.App{ID: ID, Name: Name, Protocol: Protocol}, handleStream(n))
}

// Run sends a greeting to peerID and returns that peer's reply.
func Run(ctx context.Context, n *node.Node, peerID peer.ID) (string, error) {
	stream, err := n.OpenAppStream(ctx, peerID, ID)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	if _, err := fmt.Fprintf(stream, "Hello from %s\n", n.Name); err != nil {
		return "", err
	}

	reply, err := bufio.NewReader(stream).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply), nil
}

func handleStream(n *node.Node) network.StreamHandler {
	return func(stream network.Stream) {
		defer stream.Close()

		greeting, err := bufio.NewReader(stream).ReadString('\n')
		if err != nil {
			fmt.Printf("Test app read failed: %v\n", err)
			return
		}

		if _, err := fmt.Fprintf(stream, "Hello from %s; received %s", n.Name, strings.TrimSpace(greeting)+"\n"); err != nil {
			fmt.Printf("Test app reply failed: %v\n", err)
		}
	}
}
