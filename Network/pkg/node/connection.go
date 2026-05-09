package node

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Connect attempts to connect to a peer given its multiaddress.
func (n *Node) Connect(addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("invalid multiaddr: %w", err)
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("invalid peer info: %w", err)
	}

	if n.Host.Network().Connectedness(info.ID) == network.Connected {
		fmt.Println("Already connected to:", info.ID)
		return nil
	}

	ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
	defer cancel()

	if err := n.Host.Connect(ctx, *info); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	return nil
}

// Disconnect attempts to disconnect from a peer given its peer ID.
func (n *Node) Disconnect(peerIDStr string) error {
	pid, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	conns := n.Host.Network().ConnsToPeer(pid)

	for _, c := range conns {
		c.Close()
	}

	n.peerMu.Lock()
	delete(n.Peers, pid)
	n.peerMu.Unlock()

	n.logf("Disconnected from: %s", pid)

	return nil
}
