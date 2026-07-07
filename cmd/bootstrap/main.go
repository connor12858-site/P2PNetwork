package main

import (
	"bytes"
	"encoding/json"
	"fgov/network/pkg/node"
	"fgov/network/pkg/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// global variables
var cfg *util.Config
var n *node.Node
var err error

// run_bootstrap uploads the bootstrap multiaddr to the API endpoint for other nodes to fetch.
func run_bootstrap() {
	// Print the bootstrap node's multiaddr for other nodes to connect to.
	fmt.Println("\nBOOTSTRAP INFO\n========")
	fmt.Println("Bootstrap node started.")

	// Prefer a non-loopback, non-link-local address for remote peers.
	address := n.Host.Addrs()[0].String()

	var externalAddress string
	for _, addr := range n.Host.Addrs() {
		addrStr := addr.String()
		// Skip loopback and link-local addresses that remote peers cannot route to.
		if !strings.Contains(addrStr, "127.0.0.1") &&
			!strings.Contains(addrStr, "::1") &&
			!strings.Contains(addrStr, "0.0.") &&
			!strings.Contains(addrStr, "169.254.") &&
			!strings.Contains(addrStr, "fe80:") {
			externalAddress = addrStr
			break
		}
	}

	// Update the address
	if externalAddress != "" {
		address = externalAddress
	}
	fmt.Println("Using external address:", address)

	node := map[string]any{
		"peer_id": n.Host.ID().String(),
		"address": address,
		"port":    cfg.Port,
		"name":    cfg.Name,
	}

	jsonData, err := json.Marshal(node)
	if err != nil {
		fmt.Println("Error marshaling bootstrap node:", err)
		util.StopProgram(1)
	}

	url := fmt.Sprintf("%s/register?password=%s", cfg.Server, cfg.Password)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		util.StopProgram(1)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error uploading bootstrap node:", err)
		util.StopProgram(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Bootstrap upload failed:", resp.Status, string(body))
		util.StopProgram(1)
	}

	fmt.Println("Bootstrap node uploaded successfully.")
}

// init loads the configuration from boot-config.yaml and prints the values for debugging.
func init() {
	// Load the config
	cfg, err = util.LoadConfig("boot-config.yaml")
	if err != nil {
		fmt.Println(err)
		util.StopProgram(1)
	}

	cfg.PrintData()
}

// main launches the bootstrap node, registers it with the server, and handles graceful shutdown.
func main() {
	// Create a new node for bootstrap
	n, err = node.NewNode(cfg.Port, nil, cfg.Name)
	if err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // kill
		syscall.SIGHUP,  // terminal closed (Unix)
	)

	go func() {

		sig := <-sigChan
		fmt.Printf("Received signal: %v\n", sig)

		peerID := n.Host.ID().String()

		url := fmt.Sprintf("%s/register/%s?password=%s",
			cfg.Server,
			peerID,
			cfg.Password,
		)

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Println("Error creating delete request:", err)
			os.Exit(1)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error deleting bootstrap node:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Println("Delete failed:", resp.Status, string(body))
		}

		fmt.Println("Bootstrap node removed successfully.")
		os.Exit(0)
	}()

	run_bootstrap()

	select {}
}
