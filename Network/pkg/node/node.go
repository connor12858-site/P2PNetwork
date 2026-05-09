package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	routingdisc "github.com/libp2p/go-libp2p/p2p/discovery/routing"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a libp2p node with its host, DHT, and peer management.
type Node struct {
    Host   host.Host
    Ctx    context.Context
    Cancel context.CancelFunc

    DHT       *dht.IpfsDHT
    Discovery discovery.Discovery

    Peers           map[peer.ID]peer.AddrInfo
    DiscoveredPeers map[peer.ID]peer.AddrInfo

    peerMu  sync.Mutex
    printMu sync.Mutex

    Name string
}

// NewNode creates and initializes a new libp2p node listening on the specified port.
func NewNode(listenPort int, bootstrapPeers []string) (*Node, error) {
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

    node := &Node{
        Host:            h,
        Ctx:             ctx,
        Cancel:          cancel,
        DHT:             kademliaDHT,
        Discovery:       routingDiscovery,
        Peers:           make(map[peer.ID]peer.AddrInfo),
        DiscoveredPeers: make(map[peer.ID]peer.AddrInfo),
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
            node.Peers[peerID] = peer.AddrInfo{
                ID:    peerID,
                Addrs: addrs,
            }
            node.peerMu.Unlock()
        },

        DisconnectedF: func(net network.Network, c network.Conn) {
            peerID := c.RemotePeer()

            node.logf("Disconnected from: %s", peerID)

            node.peerMu.Lock()
            delete(node.Peers, peerID)
            node.peerMu.Unlock()
        },
    })

    return node, nil
}

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

// Close gracefully shuts down the node by canceling its context and closing the host.
func (n *Node) Close() {
    n.Cancel()
    n.Host.Close()
}

// PrintPeers displays the list of currently connected peers.
func (n *Node) PrintPeers() {
    n.peerMu.Lock()
    defer n.peerMu.Unlock()

    fmt.Println("Connected peers:")

    for id := range n.Peers {
        fmt.Println("-", id)
    }
}

// PrintDiscoveredPeers displays discovered peers.
func (n *Node) PrintDiscoveredPeers() {
    n.peerMu.Lock()
    defer n.peerMu.Unlock()

    fmt.Println("Discovered peers:")

    for id := range n.DiscoveredPeers {
        fmt.Println("-", id)
    }
}

// logf is a helper function to print messages while trying to keep the command prompt clean.
func (n *Node) logf(format string, args ...interface{}) {
    n.printMu.Lock()
    defer n.printMu.Unlock()

    fmt.Print("\r\033[K")

    fmt.Printf(format+"\n", args...)

    fmt.Print("> ")
}

// Logf exposes the node logger for app packages that share the same terminal.
func (n *Node) Logf(format string, args ...interface{}) {
    n.logf(format, args...)
}

// waitForRoutingPeers blocks until the DHT routing table contains at least minPeers peers.
func (n *Node) waitForRoutingPeers(minPeers int, timeout time.Duration) error {
    deadline := time.NewTimer(timeout)
    defer deadline.Stop()

    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    refreshTicker := time.NewTicker(5 * time.Second)
    defer refreshTicker.Stop()

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

// FindPeers runs discovery and connects to discovered peers (exported for use by callers).
func (n *Node) FindPeers() {
    time.Sleep(3 * time.Second)

    for {
        peerChan, err := n.Discovery.FindPeers(n.Ctx, "fgov-network")
        if err != nil {
            time.Sleep(5 * time.Second)
            continue
        }

        for p := range peerChan {
            if p.ID == n.Host.ID() {
                continue
            }

            n.peerMu.Lock()
            n.DiscoveredPeers[p.ID] = p
            n.peerMu.Unlock()

            if n.Host.Network().Connectedness(p.ID) == network.Connected {
                continue
            }

            go func(p peer.AddrInfo) {
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

// GetPeersSnapshot returns a slice of peer.AddrInfo for currently connected peers.
func (n *Node) GetPeersSnapshot() []peer.AddrInfo {
    n.peerMu.Lock()
    defer n.peerMu.Unlock()

    out := make([]peer.AddrInfo, 0, len(n.Peers))
    for _, info := range n.Peers {
        out = append(out, info)
    }
    return out
}

// GetDiscoveredPeersSnapshot returns a slice of peer.AddrInfo for discovered peers.
func (n *Node) GetDiscoveredPeersSnapshot() []peer.AddrInfo {
    n.peerMu.Lock()
    defer n.peerMu.Unlock()

    out := make([]peer.AddrInfo, 0, len(n.DiscoveredPeers))
    for _, info := range n.DiscoveredPeers {
        out = append(out, info)
    }
    return out
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

// StartDiscovery launches the discovery advertise and find loops in background goroutines.
// It is safe to call multiple times; subsequent calls will start additional goroutines.
func (n *Node) StartDiscovery(topic string) {
    go n.AdvertiseDiscovery(topic)
    go n.FindPeers()
}
