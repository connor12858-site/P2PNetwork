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
	"syscall"
)

var cfg *util.Config
var n *node.Node
var err error

// Initialize the bootstrap node and print its multiaddr for other nodes to connect to.
func init() {
	cfg, err = util.LoadConfig("boot-config.yaml")
	if err != nil {
		fmt.Println(err)
		util.StopProgram(1)
	}

	fmt.Println(cfg.Port)
	fmt.Println(cfg.Topic)
	fmt.Println(cfg.Server)
	fmt.Println(cfg.Logging)
	fmt.Println(cfg.Name)

	n, err = node.NewNode(cfg.Port, nil, cfg.Name)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nBootstrap node started.")

	// Print the peer ID to let the user copy the bootstrap multiaddr if needed.
	fmt.Println("Peer ID:", n.Host.ID().String())

	save_addr := n.Host.Addrs()[0].String() + "/p2p/" + n.Host.ID().String()
	fmt.Println("Bootstrap multiaddr:", save_addr)
	fmt.Println()
}

// run_bootstrap uploads the bootstrap multiaddr to the API endpoint for other nodes to fetch.
func run_bootstrap() {
	address := n.Host.Addrs()[0].String()

	node := map[string]string{
		"peer_id": n.Host.ID().String(),
		"address": address,
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

func main() {
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
