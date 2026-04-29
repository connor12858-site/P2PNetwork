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
}

// NewNode creates and initializes a new libp2p node listening on the specified port.
func NewNode(listenPort int) (*Node, error) {
	// Create the context and cancel function for the node
	ctx, cancel := context.WithCancel(context.Background())

	// Open a port for listening to incoming connections
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
		),
	)
	if err != nil {
		cancel()
		return nil, err
	}

	// Initialize the Kademlia DHT for peer discovery
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		cancel()
		return nil, err
	}

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		cancel()
		return nil, err
	}

	// Connect to bootstrap peers to join the network
	bootstrapPeers := []string{
		// "/ip4/10.0.0.3/tcp/57866/p2p/12D3KooWPYV51GCg5b98P8HrN2vHddQLe4BRwEiCmedJbrTLW4Yw",
	}

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

		if err := h.Connect(ctx, *info); err != nil {
			fmt.Println("Bootstrap connect failed:", err)
		} else {
			fmt.Println("Connected to bootstrap:", info.ID)
		}
	}

	routingDiscovery := routingdisc.NewRoutingDiscovery(kademliaDHT)

	// Create the node that the main function will interact with
	node := &Node{
		Host:            h,
		Ctx:             ctx,
		Cancel:          cancel,
		DHT:             kademliaDHT,
		Discovery:       routingDiscovery,
		Peers:           make(map[peer.ID]peer.AddrInfo),
		DiscoveredPeers: make(map[peer.ID]peer.AddrInfo),
	}

	// Log the node's ID and listening addresses
	fmt.Println("Node started")
	fmt.Println("Peer ID:", h.ID().String())
	fmt.Println("Full addresses:")

	for _, addr := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, h.ID())
	}

	// Set up network notifications to track peer connections and disconnections
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
	// Parse the multiaddress and extract peer information
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("invalid multiaddr: %w", err)
	}

	// Convert the multiaddress to peer information
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("invalid peer info: %w", err)
	}

	// Check if we're already connected to this peer
	if n.Host.Network().Connectedness(info.ID) == network.Connected {
		fmt.Println("Already connected to:", info.ID)
		return nil
	}

	// Attempt to connect to the peer with a timeout
	ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
	defer cancel()

	if err := n.Host.Connect(ctx, *info); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	return nil
}

// Disconnect attempts to disconnect from a peer given its peer ID.
func (n *Node) Disconnect(peerIDStr string) error {
	// Decode the peer ID from the provided string
	pid, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Close all connections to the specified peer
	conns := n.Host.Network().ConnsToPeer(pid)

	for _, c := range conns {
		c.Close()
	}

	// Remove the peer from our internal tracking map
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

	// Clear current line
	fmt.Print("\r\033[K")

	fmt.Printf(format+"\n", args...)

	// redraw prompt
	fmt.Print("> ")
}

func (n *Node) findPeers() {
	// Give DHT time to warm up
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

			// Skip already connected peers
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

// main is the entry point of the application, providing a simple command-line interface for interacting with the libp2p node.
func main() {
	// Create a new node listening on port 0 (random available port)
	node, err := NewNode(0)
	if err != nil {
		panic(err)
	}

	// Advertise the node on the "fgov-network" topic for peer discovery
	_, err = node.Discovery.Advertise(node.Ctx, "fgov-network")
	if err != nil {
		panic(err)
	}

	go node.findPeers()

	// Set up a scanner to read user input from the command line
	scanner := bufio.NewScanner(os.Stdin)

	// Print available commands to the user
	fmt.Println("\nCommands:")
	fmt.Println(" connect <multiaddr>")
	fmt.Println(" disconnect <peerID>")
	fmt.Println(" peers")
	fmt.Println(" discovered")
	fmt.Println(" exit")

	node.printMu.Lock()
	fmt.Print("> ")
	node.printMu.Unlock()

	// Main loop to process user commands
	for {
		// Get the next line of input from the user
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		parts := strings.SplitN(input, " ", 2)
		cmd := parts[0]

		// Handle the command based on user input
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

		case "discovered":
			node.PrintDiscoveredPeers()

		case "exit":
			node.Close()
			return

		default:
			fmt.Println("Unknown command")
		}

		node.printMu.Lock()
		fmt.Print("> ")
		node.printMu.Unlock()
	}
}
