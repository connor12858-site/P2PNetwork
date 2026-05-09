package main

import (
	"fmt"
	"os"

	"fgov/network/pkg/node"
)

const PORT = 52837
const TOPIC = "fgov-network"

func main() {
    var name string
	fmt.Print("Enter a name for this bootstrap node: ")
	fmt.Scanln(&name)
	name += "-bootstrap"

    n, err := node.NewNode(PORT, nil, name)
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
