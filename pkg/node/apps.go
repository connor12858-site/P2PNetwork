package node

import (
	"context"
	"fmt"
	"sort"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// App describes a peer-to-peer feature exposed by a node.
// ID is used to open the app; Protocol identifies its libp2p stream protocol.
type App struct {
	ID       string
	Name     string
	Protocol protocol.ID
}

// RegisterApp makes an app available to remote peers and to the local GUI.
// An app ID and protocol may each be registered only once for a node.
func (n *Node) RegisterApp(app App, handler network.StreamHandler) error {
	if app.ID == "" || app.Name == "" || app.Protocol == "" {
		return fmt.Errorf("app ID, name, and protocol are required")
	}
	if handler == nil {
		return fmt.Errorf("app %q requires a stream handler", app.ID)
	}

	n.appMu.Lock()
	defer n.appMu.Unlock()

	if _, exists := n.apps[app.ID]; exists {
		return fmt.Errorf("app %q is already registered", app.ID)
	}
	for _, registered := range n.apps {
		if registered.Protocol == app.Protocol {
			return fmt.Errorf("protocol %q is already registered", app.Protocol)
		}
	}

	n.Host.SetStreamHandler(app.Protocol, handler)
	n.apps[app.ID] = app
	return nil
}

// RegisteredApps returns a stable snapshot suitable for rendering in a UI.
func (n *Node) RegisteredApps() []App {
	n.appMu.RLock()
	defer n.appMu.RUnlock()

	apps := make([]App, 0, len(n.apps))
	for _, app := range n.apps {
		apps = append(apps, app)
	}
	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })
	return apps
}

// OpenAppStream opens a stream for a locally registered app.
func (n *Node) OpenAppStream(ctx context.Context, peerID peer.ID, appID string) (network.Stream, error) {
	n.appMu.RLock()
	app, exists := n.apps[appID]
	n.appMu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("app %q is not registered", appID)
	}

	return n.Host.NewStream(ctx, peerID, app.Protocol)
}
