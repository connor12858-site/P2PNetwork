package node

import (
	"context"
	"fmt"
	"strings"
	"sync"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	routingdisc "github.com/libp2p/go-libp2p/p2p/discovery/routing"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const peerNameProtocolID protocol.ID = "/fgov"

// PeerRecord stores a peer identity together with a display name.
type PeerRecord struct {
	Name     string
	AddrInfo peer.AddrInfo
}

// Node represents a libp2p node with its host, DHT, and peer management.
type Node struct {
	Host   host.Host
	Ctx    context.Context
	Cancel context.CancelFunc
	Name   string

	DHT       *dht.IpfsDHT
	Discovery discovery.Discovery

	Peers           map[peer.ID]PeerRecord
	DiscoveredPeers map[peer.ID]PeerRecord

	peerMu  sync.Mutex
	printMu sync.Mutex
}

// NewNode creates and initializes a new libp2p node listening on the specified port.
func NewNode(listenPort int, bootstrapPeers []string, name string) (*Node, error) {
	// Create a cancellable context for the node's lifecycle.
	ctx, cancel := context.WithCancel(context.Background())

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
		),
		libp2p.NATPortMap(),
		libp2p.ResourceManager(&network.NullResourceManager{}), // Disable resource manager to avoid connection limits
	)
	if err != nil {
		cancel()

		return nil, err
	}

	// Convert bootstrap peer multiaddrs to AddrInfo for DHT configuration and connection.
	bootstrapAddrInfos := make([]peer.AddrInfo, 0, len(bootstrapPeers))
	for _, addr := range bootstrapPeers {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			fmt.Println("Invalid bootstrap multiaddr:", err)
			continue
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			fmt.Println("Invalid bootstrap peer info:", err)
			continue
		}

		bootstrapAddrInfos = append(bootstrapAddrInfos, *info)
	}

	// Set up the DHT in server mode, optionally seeding with bootstrap peers for faster routing table population.
	dhtOptions := make([]dht.Option, 0, 2)
	dhtOptions = append(dhtOptions, dht.Mode(dht.ModeServer))
	if len(bootstrapAddrInfos) > 0 {
		// Seed DHT bootstrap with the configured peers so routing table refresh can populate.
		dhtOptions = append(dhtOptions, dht.BootstrapPeers(bootstrapAddrInfos...))
	}

	// Create the Kademlia DHT and bootstrap it to populate the routing table.
	kademliaDHT, err := dht.New(ctx, h, dhtOptions...)
	if err != nil {
		cancel()
		return nil, err
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		cancel()
		return nil, err
	}

	// Connect to bootstrap peers to populate routing table immediately (optional but helps with faster discovery).
	for _, info := range bootstrapAddrInfos {
		if err := h.Connect(ctx, info); err != nil {
			fmt.Println("Bootstrap connect failed:", err)
		} else {
			fmt.Println("Connected to bootstrap:", info.ID)
		}
	}

	// Set up the routing discovery service using the DHT for peer discovery.
	routingDiscovery := routingdisc.NewRoutingDiscovery(kademliaDHT)

	// Trim whitespace from the provided name and create the Node struct with initialized fields.
	nodeName := strings.TrimSpace(name)
	node := &Node{
		Host:            h,
		Ctx:             ctx,
		Cancel:          cancel,
		Name:            nodeName,
		DHT:             kademliaDHT,
		Discovery:       routingDiscovery,
		Peers:           make(map[peer.ID]PeerRecord),
		DiscoveredPeers: make(map[peer.ID]PeerRecord),
	}

	// Log the node startup information, including peer ID and listening addresses.
	fmt.Println("\nNODE INFO\n========")
	fmt.Println("Node started")
	fmt.Println("Peer ID:", h.ID().String())

	fmt.Println("\nFull addresses:")
	for _, addr := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, h.ID())
	}

	// Set up network notifications to track peer connections and disconnections, updating the Peers map accordingly.
	h.Network().Notify(&network.NotifyBundle{
		// When a new peer connection is established, add the peer to the Peers map and log the connection.
		ConnectedF: func(net network.Network, c network.Conn) {
			peerID := c.RemotePeer()

			node.logf("Connected to: %s", peerID)

			addrs := node.Host.Peerstore().Addrs(peerID)

			node.peerMu.Lock()
			node.Peers[peerID] = PeerRecord{
				Name:     defaultPeerName(peerID.String()),
				AddrInfo: peer.AddrInfo{ID: peerID, Addrs: addrs},
			}
			node.peerMu.Unlock()

			go node.syncPeerName(peerID)
		},

		// When a peer disconnects, remove the peer from the Peers and DiscoveredPeers maps and log the disconnection.
		DisconnectedF: func(net network.Network, c network.Conn) {
			peerID := c.RemotePeer()

			node.logf("Disconnected from: %s", peerID)

			node.peerMu.Lock()
			delete(node.Peers, peerID)
			delete(node.DiscoveredPeers, peerID)
			node.peerMu.Unlock()
		},
	})

	// Set a stream handler for the peer name protocol to handle incoming requests for peer names.
	h.SetStreamHandler(peerNameProtocolID, node.handlePeerNameStream)

	return node, nil
}
