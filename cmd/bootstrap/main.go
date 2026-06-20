package main

import (
	"bytes"
	"fgov/network/pkg/node"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const PORT = 52837
const TOPIC = "fgov-network"

const BOOTSTRAP_API = "https://fgov.connor12858.workers.dev"

// const BOOTSTRAP_API = "http://127.0.0.1:8787/"

var n *node.Node
var err error

// Check if the machine is connected to the internet by making a simple HTTP request.
func isConnected() bool {
	timeout := 5 * time.Second
	client := http.Client{Timeout: timeout}
	_, err := client.Get("https://firefox.com")
	return err == nil
}

// Initialize the bootstrap node and print its multiaddr for other nodes to connect to.
func init() {
	var name string
	fmt.Print("Enter a name for this bootstrap node: ")
	fmt.Scanln(&name)
	name += "-bootstrap"

	n, err = node.NewNode(PORT, nil, name)
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

// If the machine is online, upload the bootstrap multiaddr to the API endpoint for other nodes to fetch. If offline, just print the multiaddr and wait for the user to copy it.
func online() {
	// Upload the bootstrap multiaddr to api end point for other nodes to fetch
	fmt.Println("Connected to the internet.")

	// Send the bootstrap multiaddr to the API endpoint as JSON
	// Create the JSON payload
	payload := fmt.Sprintf(`{"name": "%s", "addr": "%s"}`, n.Name, n.Host.Addrs()[0].String())
	fmt.Println(payload)
	// Send the POST request
	resp, err := http.Post(BOOTSTRAP_API, "application/json", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		fmt.Println("Error uploading bootstrap multiaddr:", err)
	} else {
		fmt.Println("Bootstrap multiaddr uploaded successfully.")
		resp.Body.Close()
	}

}

func offline() {
	fmt.Println("Not connected to the internet.")
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

		// Delete the bootstrap multiaddr from the API endpoint if online
		if isConnected() {
			payload := fmt.Sprintf(`{"delete": "%s"}`, n.Name)
			req, err := http.Post(BOOTSTRAP_API, "application/json", bytes.NewBuffer([]byte(payload)))
			if err != nil {
				fmt.Println("Error deleting bootstrap multiaddr:", err)
			} else {
				fmt.Println("Bootstrap multiaddr deleted successfully.")
				req.Body.Close()
			}
		} else {
			fmt.Println("Not connected to the internet, skipping deletion of bootstrap multiaddr.")
		}

		os.Exit(0)
	}()

	// Detect if connected to internet
	if isConnected() {
		online()
	} else {
		offline()
	}

	select {}
}
