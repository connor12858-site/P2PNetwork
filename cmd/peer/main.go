package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"fgov/network/pkg/node"
)

// global variables
var PORT = 0;
var TOPIC = ""
var BOOTSTRAP_ADDRESS = ""
var BOOTSTRAP_PEERS = make([]string, 0, 1)

// config reads the configuration from config.yaml and sets the global variables.
func config() {
    // Check for the yaml file first
    if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
        fmt.Println("config.yaml not found. Please create the file with the following content:")
        fmt.Println("port: <your_port>")
        fmt.Println("topic: <your_topic>")
        fmt.Println("bootstrap: <bootstrap_address>")
        os.Exit(1)
    }

    // Gather the data
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        fmt.Println("Error reading config.yaml:", err)
        os.Exit(1)
    }

    // Save the data
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "port:") {
            fmt.Sscanf(line, "port: %d", &PORT)
        } else if strings.HasPrefix(line, "topic:") {
            fmt.Sscanf(line, "topic: %s", &TOPIC)
        } else if strings.HasPrefix(line, "bootstrap:") {
            fmt.Sscanf(line, "bootstrap: %s", &BOOTSTRAP_ADDRESS)
        }
    }
}

// read the bootstrap address 
func get_bootstrap() {
    data, err := os.ReadFile("bs-nodes")
    if err == nil && len(strings.TrimSpace(string(data))) > 0 {
        bootstrap := strings.TrimSpace(string(data))
        fmt.Println("Bootstrap multiaddr from file:", bootstrap)
        BOOTSTRAP_PEERS = append(BOOTSTRAP_PEERS, bootstrap)
    }
}

// main is the entry point of the application and acts as the peer runner.
func main() {
    config()

    get_bootstrap()
   
    var name string
    fmt.Print("Enter a name for this node: ")
    fmt.Scanln(&name)

    n, err := node.NewNode(PORT, BOOTSTRAP_PEERS, name)
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
