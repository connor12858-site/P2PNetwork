package main

import (
	"flag"
	"fmt"
	"os"

	"fgov/network/pkg/node"
)

func main() {
    port := flag.Int("port", 52837, "listen port for bootstrap node")
    flag.Parse()

    n, err := node.NewNode(*port, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println("Bootstrap node started.")

    // Print the peer ID to let the user copy the bootstrap multiaddr if needed.
    fmt.Println("Peer ID:", n.Host.ID().String())

	save_addr := n.Host.Addrs()[0].String() + "/p2p/" + n.Host.ID().String()
	fmt.Println("Bootstrap multiaddr:", save_addr)

	// Create the file for the address if not there
	file, _ := os.Create("bs-nodes")
	// Write to the file
	file.WriteString(save_addr)
	// Close the file
	file.Close()

    // Keep the bootstrap node running indefinitely.
    // Optionally you can expand this to expose administrative commands.
    select {}
}
