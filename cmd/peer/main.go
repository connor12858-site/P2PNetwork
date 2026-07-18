package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"fgov/network/pkg/node"
	"fgov/network/pkg/util"
)

// global variables
type BootstrapNode struct {
	PeerID  string `json:"peer_id"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

type BootstrapResponse struct {
	Nodes []BootstrapNode `json:"nodes"`
}

var BOOTSTRAP_PEERS = make([]string, 0, 1)
var cfg *util.Config
var err error

// read the bootstrap address
func get_bootstrap() {
	resp, err := http.Get(cfg.Server)
	if err != nil {
		cfg.DebugLog("Error fetching bootstrap nodes:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		cfg.DebugLog("Bootstrap fetch failed:", resp.Status, string(body))
		return
	}

	var result BootstrapResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		cfg.DebugLog("Error decoding bootstrap response:", err)
		return
	}

	for _, node := range result.Nodes {
		if strings.TrimSpace(node.Address) == "" {
			continue
		}

		full_addr := node.Address + "/p2p/" + node.PeerID
		cfg.DebugLog("Bootstrap node:", full_addr)
		BOOTSTRAP_PEERS = append(BOOTSTRAP_PEERS, full_addr)
	}
}

// init loads the configuration from config.yaml and prints the values for debugging.
func init() {
	cfg, err = util.LoadConfig("config.yaml")
	if err != nil {
		cfg.DebugLog("Error loading config:", err)
		util.StopProgram(1)
	}

	cfg.PrintData()
}

// main is the entry point of the application and acts as the peer runner.
func main() {
	get_bootstrap()

	// Create a new node with the specified port, bootstrap peers, and name.
	n, err := node.NewNode(cfg.Port, BOOTSTRAP_PEERS, cfg.Name)
	if err != nil {
		panic(err)
	}

	n.StartDiscovery(cfg.Topic)

	// Start a command-line interface for user interaction with the node.
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

	if err := scanner.Err(); err != nil {
		fmt.Println("Scanner error:", err)
	}
}
