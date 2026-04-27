package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
    Host    host.Host
    Ctx     context.Context
    Cancel  context.CancelFunc
	
	DHT		*dht.IpfsDHT
	Discorvery discovery.Discovery

    Peers   map[peer.ID]peer.AddrInfo
	peerMu 	sync.Mutex

    printMu sync.Mutex
}

func NewNode(listenPort int) (*Node, error) {
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

	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		cancel()
		return nil, err
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

    node := &Node{
        Host:   h,
        Ctx:    ctx,
        Cancel: cancel,
        Peers:  make(map[peer.ID]peer.AddrInfo),
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

        node.logf("Connected to:", peerID)

        n := node // capture outer node

        n.peerMu.Lock()
        n.Peers[peerID] = peer.AddrInfo{
            ID:    peerID,
            Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
        }
        n.peerMu.Unlock()
    },
    DisconnectedF: func(net network.Network, c network.Conn) {
        peerID := c.RemotePeer()

        node.logf("Disconnected from:", peerID)

        n := node

        n.peerMu.Lock()
        delete(n.Peers, peerID)
        n.peerMu.Unlock()
    },
})

    return node, nil
}

func (n *Node) Connect(addr string) error {
    maddr, err := multiaddr.NewMultiaddr(addr)
     if err != nil {
        return fmt.Errorf("invalid multiaddr: %w", err)
    }

    info, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        return fmt.Errorf("invalid peer info: %w", err)
    }
	
	n.peerMu.Lock()
	if _, exists := n.Peers[info.ID]; exists {
		n.peerMu.Unlock()
		fmt.Println("Already connected to:", info.ID)
		return nil
	}
	n.peerMu.Unlock()

	ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
	defer cancel()
    if err := n.Host.Connect(ctx, *info); err != nil {
        return fmt.Errorf("connection failed: %w", err)
    }

    return nil
}

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

    n.logf("Disconnected from:", pid)
    return nil
}

func (n *Node) Close() {
    n.Cancel()
    n.Host.Close()
}

func (n *Node) PrintPeers() {
    n.peerMu.Lock()
    defer n.peerMu.Unlock()

    fmt.Println("Connected peers:")
    for id := range n.Peers {
        fmt.Println("-", id)
    }
}

func (n *Node) logf(format string, args ...interface{}) {
    n.printMu.Lock()
    defer n.printMu.Unlock()

    // Clear current line (basic attempt)
    fmt.Print("\r\033[K")

    fmt.Printf(format+"\n", args...)

    // redraw prompt
    fmt.Print("> ")
}

func main() {
    node, err := NewNode(0)
    if err != nil {
        panic(err)
    }

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Println("\nCommands:")
    fmt.Println(" connect <multiaddr>")
    fmt.Println(" disconnect <peerID>")
    fmt.Println(" peers")
    fmt.Println(" exit")

    for {
		node.printMu.Lock()
		fmt.Print("> ")
		node.printMu.Unlock()
        if !scanner.Scan() {
            break
        }

        input := scanner.Text()
        parts := strings.SplitN(input, " ", 2)

        cmd := parts[0]

        switch cmd {

        case "connect":
            if len(parts) < 2 {
                fmt.Println("Usage: connect <multiaddr>")
                continue
            }
            if err := node.Connect(parts[1]); err != nil {
                fmt.Println("Error:", err)
            }

        case "disconnect":
            if len(parts) < 2 {
                fmt.Println("Usage: disconnect <peerID>")
                continue
            }
            if err := node.Disconnect(parts[1]); err != nil {
                fmt.Println("Error:", err)
            }

        case "peers":
            node.PrintPeers()

        case "exit":
            node.Close()
            return

        default:
            fmt.Println("Unknown command")
        }
    }
}