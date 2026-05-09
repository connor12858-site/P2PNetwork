package node

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type peerNamePayload struct {
	Name string `json:"name"`
}

// defaultPeerName generates a default name for a peer based on its ID.
func defaultPeerName(peerID string) string {
	return "peer-" + shortPeerLabel(peerID)
}

// shortPeerLabel returns the first 8 characters of a peer ID, or the full ID if shorter.
func shortPeerLabel(peerID string) string {
	if len(peerID) <= 8 {
		return peerID
	}
	return peerID[:8]
}

// updatePeerName updates the name for a peer in the peer map.
func (n *Node) updatePeerName(peerID peer.ID, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	n.peerMu.Lock()
	defer n.peerMu.Unlock()

	record, ok := n.Peers[peerID]
	if !ok {
		record = PeerRecord{AddrInfo: peer.AddrInfo{ID: peerID}}
	}
	record.Name = name
	n.Peers[peerID] = record
}

// displayName returns the node's display name.
func (n *Node) displayName() string {
	return n.Name
}

// Creates a default node name based on the hostname and peer ID.
func defaultNodeName(peerID string) string {
    hostname, err := os.Hostname()
    if err != nil || strings.TrimSpace(hostname) == "" {
        return "node-" + shortPeerLabel(peerID)
    }
    return strings.TrimSpace(hostname) + "-" + shortPeerLabel(peerID)
}

// syncPeerName exchanges names with a peer and updates the peer's name in the local peer map.
func (n *Node) syncPeerName(peerID peer.ID) {
	ctx, cancel := context.WithTimeout(n.Ctx, 5*time.Second)
	defer cancel()

	stream, err := n.Host.NewStream(ctx, peerID, peerNameProtocolID)
	if err != nil {
		return
	}
	defer stream.Close()

	encoder := json.NewEncoder(stream)
	decoder := json.NewDecoder(stream)

	if err := encoder.Encode(peerNamePayload{Name: n.displayName()}); err != nil {
		return
	}

	var payload peerNamePayload
	if err := decoder.Decode(&payload); err != nil {
		return
	}

	n.updatePeerName(peerID, payload.Name)
}

// handlePeerNameStream handles incoming peer name exchange requests.
func (n *Node) handlePeerNameStream(stream network.Stream) {
	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()
	decoder := json.NewDecoder(bufio.NewReader(stream))
	encoder := json.NewEncoder(stream)

	var payload peerNamePayload
	if err := decoder.Decode(&payload); err != nil {
		return
	}

	if err := encoder.Encode(peerNamePayload{Name: n.displayName()}); err != nil {
		return
	}

	n.updatePeerName(remotePeerID, payload.Name)
}
