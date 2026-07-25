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
	"time"
)

// global variables
var cfg *util.Config
var n *node.Node
var err error

type BootstrapNode struct {
	PeerID  string `json:"peer_id"`
	Address string `json:"address"`
}

type BootstrapResponse struct {
	Nodes []BootstrapNode `json:"nodes"`
}

// getBootstrapPeers returns bootstrap nodes already registered with the
// registry. Joining them keeps all bootstrap nodes in one DHT routing domain.
func getBootstrapPeers() []string {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(cfg.Server)
	if err != nil {
		cfg.DebugLog("Error fetching existing bootstrap nodes:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		cfg.DebugLog("Bootstrap fetch failed:", resp.Status, string(body))
		return nil
	}

	var result BootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		cfg.DebugLog("Error decoding bootstrap response:", err)
		return nil
	}

	peers := make([]string, 0, len(result.Nodes))
	seen := make(map[string]struct{}, len(result.Nodes))
	for _, bootstrap := range result.Nodes {
		address := strings.TrimSpace(bootstrap.Address)
		peerID := strings.TrimSpace(bootstrap.PeerID)
		if address == "" || peerID == "" {
			continue
		}

		fullAddr := address + "/p2p/" + peerID
		if _, ok := seen[fullAddr]; ok {
			continue
		}
		seen[fullAddr] = struct{}{}
		peers = append(peers, fullAddr)
	}

	return peers
}

// run_bootstrap uploads the bootstrap multiaddr to the API endpoint for other nodes to fetch.
func run_bootstrap() {
	// Print the bootstrap node's multiaddr for other nodes to connect to.
	cfg.DebugLog("\nBOOTSTRAP INFO\n========")
	cfg.DebugLog("Bootstrap node started.")

	cfg.DebugLog("Addresses:")
	for _, addr := range n.Host.Addrs() {
		cfg.DebugLog(addr)
	}

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
	cfg.DebugLog("Using external address:", address)

	node := map[string]any{
		"peer_id": n.Host.ID().String(),
		"address": address,
		"port":    cfg.Port,
		"name":    cfg.Name,
	}

	jsonData, err := json.Marshal(node)
	if err != nil {
		cfg.DebugLog("Error marshaling bootstrap node:", err)
		util.StopProgram(1)
	}

	url := fmt.Sprintf("%s/register?password=%s", cfg.Server, cfg.Password)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		cfg.DebugLog("Error creating request:", err)
		util.StopProgram(1)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cfg.DebugLog("Error uploading bootstrap node:", err)
		util.StopProgram(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		cfg.DebugLog("Bootstrap upload failed:", resp.Status, string(body))
		util.StopProgram(1)
	}

	cfg.DebugLog("Bootstrap node uploaded successfully.")
}

// init loads the configuration from config.yaml and prints the values for debugging.
func init() {
	// Load the config
	cfg, err = util.LoadConfig("config.yaml")
	if err != nil {
		cfg.DebugLog("Error loading config:", err)
		util.StopProgram(1)
	}

	cfg.PrintData()
}

// main launches the bootstrap node, registers it with the server, and handles graceful shutdown.
func main() {
	// Join existing bootstrap nodes before registering this one so they share a DHT.
	n, err = node.NewNode(cfg.Port, getBootstrapPeers(), cfg.Name+"-bootstrap")
	if err != nil {
		cfg.DebugLog("Error creating bootstrap node:", err)
		util.StopProgram(1)
	}

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // kill
		syscall.SIGHUP,  // terminal closed (Unix)
	)

	go func() {

		sig := <-sigChan
		cfg.DebugLog("Received signal: %v\n", sig)

		peerID := n.Host.ID().String()

		url := fmt.Sprintf("%s/register/%s?password=%s",
			cfg.Server,
			peerID,
			cfg.Password,
		)

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			cfg.DebugLog("Error creating delete request:", err)
			os.Exit(1)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			cfg.DebugLog("Error deleting bootstrap node:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			body, _ := io.ReadAll(resp.Body)
			cfg.DebugLog("Delete failed:", resp.Status, string(body))
		}

		cfg.DebugLog("Bootstrap node removed successfully.")
		os.Exit(0)
	}()

	run_bootstrap()

	select {}
}
