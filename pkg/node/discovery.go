package node

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// FindPeers runs discovery for topic and connects to discovered peers.
func (n *Node) FindPeers(topic string) {
	time.Sleep(3 * time.Second)

	// Keep retrying discovery until the DHT has peers available.
	for {
		// Wait for at least one routing peer before attempting discovery.
		peerChan, err := n.Discovery.FindPeers(n.Ctx, topic)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// Process discovered peers
		for p := range peerChan {
			if p.ID == n.Host.ID() {
				continue
			}

			// Add the discovered peer to the node's peer list if not already present.
			n.peerMu.Lock()
			n.DiscoveredPeers[p.ID] = PeerRecord{
				Name:     defaultPeerName(p.ID.String()),
				AddrInfo: p,
			}
			n.peerMu.Unlock()

			if n.Host.Network().Connectedness(p.ID) == network.Connected {
				continue
			}

			// Attempt to connect to the discovered peer in a separate goroutine.
			var dialSem = make(chan struct{}, 5) // Limit concurrent dials to 5
			go func(p peer.AddrInfo) {
				dialSem <- struct{}{} // Acquire a slot
				defer func() { <-dialSem }()

				ctx, cancel := context.WithTimeout(n.Ctx, 5*time.Second)
				defer cancel()

				if err := n.Host.Connect(ctx, p); err != nil {
					n.logf("Failed to connect to %s: %v", p.ID, err)
					return
				}

				n.logf("Discovered and connected to: %s", p.ID)
			}(p)
		}

		time.Sleep(10 * time.Second)
	}
}

// AdvertiseDiscovery keeps retrying discovery advertisement until the DHT has peers available.
func (n *Node) AdvertiseDiscovery(topic string) {
	for {
		if err := n.waitForRoutingPeers(1, 20*time.Second); err != nil {
			n.logf("Discovery advertise deferred: %v", err)

			select {
			case <-time.After(5 * time.Second):
				continue
			case <-n.Ctx.Done():
				return
			}
		}

		if _, err := n.Discovery.Advertise(n.Ctx, topic); err == nil {
			n.logf("Advertised discovery topic: %s", topic)
			return
		} else {
			n.logf("Discovery advertise failed: %v", err)
		}

		select {
		case <-time.After(5 * time.Second):
		case <-n.Ctx.Done():
			return
		}
	}
}

// waitForRoutingPeers blocks until the DHT routing table contains at least minPeers peers.
func (n *Node) waitForRoutingPeers(minPeers int, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	refreshTicker := time.NewTicker(5 * time.Second)
	defer refreshTicker.Stop()

	// Keep checking the routing table until we have enough peers or timeout occurs.
	for {
		if len(n.DHT.RoutingTable().ListPeers()) >= minPeers {
			return nil
		}

		select {
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for DHT routing peers")
		case <-n.Ctx.Done():
			return n.Ctx.Err()
		case <-ticker.C:
		case <-refreshTicker.C:
			if len(n.Host.Network().Peers()) > 0 {
				// Trigger a refresh pass when we have transport connections but no RT peers yet.
				ch := n.DHT.ForceRefresh()
				go func() {
					select {
					case <-ch:
					case <-n.Ctx.Done():
					}
				}()
			}
		}
	}
}

// StartDiscovery launches the discovery advertise and find loops in background goroutines.
// It is safe to call multiple times; subsequent calls will start additional goroutines.
func (n *Node) StartDiscovery(topic string) {
	go n.AdvertiseDiscovery(topic)
	go n.FindPeers(topic)
}
