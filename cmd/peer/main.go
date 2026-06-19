package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"fgov/network/pkg/node"
)

const PORT = 0;
const TOPIC = "fgov-network"

// main is the entry point of the application and acts as the peer runner.
func main() {
    bootstrapPeers := make([]string, 0, 1)
    data, err := os.ReadFile("bs-nodes")
    if err == nil && len(strings.TrimSpace(string(data))) > 0 {
        bootstrap := strings.TrimSpace(string(data))
        fmt.Println("Bootstrap multiaddr from file:", bootstrap)
        bootstrapPeers = append(bootstrapPeers, bootstrap)
    }

    var name string
    fmt.Print("Enter a name for this node: ")
    fmt.Scanln(&name)

    n, err := node.NewNode(PORT, bootstrapPeers, name)
    if err != nil {
        panic(err)
    }

    go n.AdvertiseDiscovery(TOPIC)
	go n.FindPeers()

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Println("\nCommands:")
    fmt.Println(" connect <multiaddr>")
    fmt.Println(" peers")
    fmt.Println(" discovered")
    fmt.Println(" exit")

    // Draw initial prompt
    fmt.Print("> ")

    for scanner.Scan() {
        input := scanner.Text()
        parts := strings.SplitN(input, " ", 3)
        cmd := parts[0]

        switch cmd {
        case "connect":
            if len(parts) < 2 {
                fmt.Println("Usage: connect <multiaddr>")
                break
            }
            if err := n.Connect(parts[1]); err != nil {
                fmt.Println("Error:", err)
            }

        case "peers":
            n.PrintPeers()

        case "discovered":
            n.PrintDiscoveredPeers()

        case "exit":
            n.Close()
            return

        default:
            fmt.Println("Unknown command")
        }

        fmt.Print("> ")
    }
}
