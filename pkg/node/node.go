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
    Name string
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
    ctx, cancel := context.WithCancel(context.Background())

    h, err := libp2p.New(
        libp2p.ListenAddrStrings(
            fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
        ),
    )
    if err != nil {
        cancel()

        return nil, err
    }

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

    dhtOptions := make([]dht.Option, 0, 2)
    dhtOptions = append(dhtOptions, dht.Mode(dht.ModeServer))
    if len(bootstrapAddrInfos) > 0 {
        // Seed DHT bootstrap with the configured peers so routing table refresh can populate.
        dhtOptions = append(dhtOptions, dht.BootstrapPeers(bootstrapAddrInfos...))
    }

    kademliaDHT, err := dht.New(ctx, h, dhtOptions...)
    if err != nil {
        cancel()
        return nil, err
    }

    if err = kademliaDHT.Bootstrap(ctx); err != nil {
        cancel()
        return nil, err
    }

    for _, info := range bootstrapAddrInfos {
        if err := h.Connect(ctx, info); err != nil {
            fmt.Println("Bootstrap connect failed:", err)
        } else {
            fmt.Println("Connected to bootstrap:", info.ID)
        }
    }

    routingDiscovery := routingdisc.NewRoutingDiscovery(kademliaDHT)

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

    fmt.Println("Node started")
    fmt.Println("Peer ID:", h.ID().String())
    fmt.Println("Full addresses:")

    for _, addr := range h.Addrs() {
        fmt.Printf("%s/p2p/%s\n", addr, h.ID())
    }

    h.Network().Notify(&network.NotifyBundle{
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

        DisconnectedF: func(net network.Network, c network.Conn) {
            peerID := c.RemotePeer()

            node.logf("Disconnected from: %s", peerID)

            node.peerMu.Lock()
            delete(node.Peers, peerID)
            delete(node.DiscoveredPeers, peerID)
            node.peerMu.Unlock()
        },
    })

    h.SetStreamHandler(peerNameProtocolID, node.handlePeerNameStream)

    return node, nil
}
