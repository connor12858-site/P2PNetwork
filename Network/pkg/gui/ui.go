package gui

import (
    "fgov/network/pkg/node"
)

// Minimal UI helpers for future expansion. Current GUI interacts directly with pkg/node.

// StartNode creates a node with the given listen port and bootstrap peers.
func StartNode(listenPort int, bootstrap []string) (*node.Node, error) {
    n, err := node.NewNode(listenPort, bootstrap)
    if err != nil {
        return nil, err
    }
    return n, nil
}
