// Package apps contains the catalogue of applications included in this build.
package apps

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"fgov/network/pkg/node"

	"github.com/libp2p/go-libp2p/core/peer"
)

// Registration describes an app that can be registered on a node and launched
// by the GUI. Apps register themselves with this catalogue from the builtin
// package, keeping the node startup code independent of individual apps.
type Registration struct {
	App      node.App
	Register func(*node.Node) error
	Run      func(context.Context, *node.Node, peer.ID) (string, error)
	Open     func(*node.Node, node.PeerRecord)
}

var (
	mu       sync.RWMutex
	registry = make(map[string]Registration)
)

// Register adds an app to the catalogue. It is intended for the compiled-in
// app catalogue during program initialization.
func Register(registration Registration) {
	if registration.App.ID == "" || registration.Register == nil || (registration.Run == nil && registration.Open == nil) {
		panic("app registration requires an ID, Register function, and Run or Open function")
	}

	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[registration.App.ID]; exists {
		panic(fmt.Sprintf("app %q is already in the catalogue", registration.App.ID))
	}
	registry[registration.App.ID] = registration
}

// All returns the compiled-in apps in display order.
func All() []Registration {
	mu.RLock()
	defer mu.RUnlock()

	registrations := make([]Registration, 0, len(registry))
	for _, registration := range registry {
		registrations = append(registrations, registration)
	}
	sort.Slice(registrations, func(i, j int) bool {
		return registrations[i].App.Name < registrations[j].App.Name
	})
	return registrations
}

// RegisterAll registers every compiled-in app on n.
func RegisterAll(n *node.Node) error {
	for _, registration := range All() {
		if err := registration.Register(n); err != nil {
			return fmt.Errorf("register %s: %w", registration.App.Name, err)
		}
	}
	return nil
}
