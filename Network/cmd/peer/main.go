package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"fgov/network/pkg/node"
)

func parseConfig() (int, []string, string) {
    port := flag.Int("port", 0, "listen port")
    bootstrap := flag.String("bootstrap", "", "bootstrap peer multiaddr for peer mode")
    topic := flag.String("topic", "fgov-network", "discovery topic")
    flag.Parse()

    bootstrapPeers := make([]string, 0, 1)
    if *bootstrap != "" {
        bootstrapPeers = append(bootstrapPeers, *bootstrap)
    }

    return *port, bootstrapPeers, *topic
}

// main is the entry point of the application and acts as the peer runner.
func main() {
    listenPort, bootstrapPeers, topic := parseConfig()
	if len(bootstrapPeers) == 0 {
        data, err := os.ReadFile("bs-nodes")
        if err == nil && len(strings.TrimSpace(string(data))) > 0 {
            bootstrap := strings.TrimSpace(string(data))
            fmt.Println("Bootstrap multiaddr from file:", bootstrap)
            bootstrapPeers = append(bootstrapPeers, bootstrap)
        }
	} 

    n, err := node.NewNode(listenPort, bootstrapPeers)
    if err != nil {
        panic(err)
    }

	go n.AdvertiseDiscovery(topic)
	go n.FindPeers()

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Println("\nCommands:")
    fmt.Println(" connect <multiaddr>")
    fmt.Println(" disconnect <peerID>")
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

        case "disconnect":
            if len(parts) < 2 {
                fmt.Println("Usage: disconnect <peerID>")
                break
            }
            if err := n.Disconnect(parts[1]); err != nil {
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
