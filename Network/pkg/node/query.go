package node

import "fmt"

// Close gracefully shuts down the node by canceling its context and closing the host.
func (n *Node) Close() {
	n.Cancel()
	n.Host.Close()
}

// IsRunning reports whether the node context is still active.
func (n *Node) IsRunning() bool {
	select {
	case <-n.Ctx.Done():
		return false
	default:
		return true
	}
}

// GetPeersSnapshot returns a slice of PeerRecord for currently connected peers.
func (n *Node) GetPeersSnapshot() []PeerRecord {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()

	out := make([]PeerRecord, 0, len(n.Peers))
	for _, info := range n.Peers {
		out = append(out, info)
	}
	return out
}

// GetDiscoveredPeersSnapshot returns a slice of PeerRecord for discovered peers.
func (n *Node) GetDiscoveredPeersSnapshot() []PeerRecord {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()

	out := make([]PeerRecord, 0, len(n.DiscoveredPeers))
	for _, info := range n.DiscoveredPeers {
		out = append(out, info)
	}
	return out
}

// PrintPeers displays the list of currently connected peers.
func (n *Node) PrintPeers() {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()

	fmt.Println("Connected peers:")

	for _, record := range n.Peers {
		fmt.Printf("- %s (%s)\n", record.Name, record.AddrInfo.ID)
	}
}

// PrintDiscoveredPeers displays discovered peers.
func (n *Node) PrintDiscoveredPeers() {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()

	fmt.Println("Discovered peers:")

	for _, record := range n.DiscoveredPeers {
		fmt.Printf("- %s (%s)\n", record.Name, record.AddrInfo.ID)
	}
}
