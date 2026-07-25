package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fgov/network/pkg/apps"
	_ "fgov/network/pkg/apps/builtin" // Load the compiled-in app catalogue.
	"fgov/network/pkg/node"
	"fgov/network/pkg/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
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

var selectedNode node.PeerRecord
var connectedPeers []node.PeerRecord
var connectedPeerNames []string

// Reads bootstrap peers from bs-nodes file.
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

func init() {
	cfg, err = util.LoadConfig("config.yaml")
	if err != nil {
		cfg.DebugLog("Error loading config:", err)
		util.StopProgram(1)
	}

	cfg.PrintData()
}

// main launches the GUI application.
func main() {
	get_bootstrap()

	a := app.New()
	w := a.NewWindow("P2P Network")
	w.Resize(fyne.NewSize(720, 480))

	var n *node.Node
	var toggle *widget.Button
	var appButtons []*widget.Button

	// Data bindings for GUI elements
	statusData := binding.NewString()
	statusData.Set("Stopped")
	nodeNameData := binding.NewString()
	nodeNameData.Set("Name: (not started)")
	progressData := binding.NewString()
	progressData.Set("Ready")
	nameData := binding.NewString()
	nameData.Set(cfg.Name)
	peerNames := binding.NewStringList()

	// GUI elements
	status := widget.NewLabelWithData(statusData)
	nodeNameLabel := widget.NewLabelWithData(nodeNameData)
	progressLabel := widget.NewLabelWithData(progressData)
	progressLabel.Wrapping = fyne.TextWrapWord

	nameEntry := widget.NewEntryWithData(nameData)
	nameEntryWrap := container.NewGridWrap(fyne.NewSize(240, nameEntry.MinSize().Height), nameEntry)
	nodeNameWrap := container.NewGridWrap(fyne.NewSize(220, nodeNameLabel.MinSize().Height), nodeNameLabel)
	statusWrap := container.NewGridWrap(fyne.NewSize(160, status.MinSize().Height), status)
	nodeNameFieldWrap := container.NewGridWrap(fyne.NewSize(90, status.MinSize().Height), widget.NewLabel("Node Name:"))

	// Create a list widget to display connected peers
	peerList := widget.NewListWithData(peerNames,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(dataItem binding.DataItem, item fyne.CanvasObject) {
			text := ""
			if value, ok := dataItem.(binding.String); ok {
				if got, err := value.Get(); err == nil {
					text = got
				}
			}
			item.(*widget.Label).SetText(text)
		},
	)

	// Update the selectedNode to be whatever node is selected in the peer list
	peerList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(connectedPeerNames) {
			peerName := connectedPeerNames[id]

			for _, p := range connectedPeers {
				if p.Name == peerName {
					selectedNode = p
					break
				}
			}

			progressData.Set(fmt.Sprintf("Selected peer: %s (%s)", selectedNode.Name, selectedNode.AddrInfo.ID))
		}
	}

	// Function to refresh the views and update the GUI elements
	refreshViews := func() {
		connectedPeerNames = make([]string, 0)

		if n != nil && n.IsRunning() {
			connectedPeers = n.GetPeersSnapshot()
			connectedPeers = util.SortPeersByName(connectedPeers)

			for _, p := range connectedPeers {
				if !strings.Contains(p.Name, "-bootstrap") {
					connectedPeerNames = append(connectedPeerNames, p.Name)
				}
			}

			statusData.Set(fmt.Sprintf("Running - %d peers", len(connectedPeers)))
			nodeNameData.Set("Name: " + n.Name)
			nameEntry.Disable()
		} else {
			statusData.Set("Stopped")
			nodeNameData.Set("Name: (not started)")
			nameEntry.Enable()
			selectedNode = node.PeerRecord{}
			peerList.UnselectAll()
		}

		if err := peerNames.Set(connectedPeerNames); err != nil {
			cfg.DebugLog("Error updating peer list:", err)
		}
		for _, appButton := range appButtons {
			if n != nil && n.IsRunning() && selectedNode.AddrInfo.ID != "" {
				appButton.Enable()
			} else {
				appButton.Disable()
			}
		}
	}

	// Start a goroutine to refresh the views every second
	go func() {
		for {
			refreshViews()
			time.Sleep(1 * time.Second)
		}
	}()

	// Create a toggle button to start/stop the node
	toggle = widget.NewButton("Turn On", func() {
		if n == nil || !n.IsRunning() {
			var err error
			progressData.Set("Starting node...")
			selectedNode = node.PeerRecord{}
			peerList.UnselectAll()
			n, err = node.NewNode(cfg.Port, BOOTSTRAP_PEERS, nameEntry.Text)
			if err != nil {
				statusData.Set("Start error: " + err.Error())
				progressData.Set("Start failed: " + err.Error())
				return
			}

			if err := apps.RegisterAll(n); err != nil {
				n.Close()
				progressData.Set("App registration failed: " + err.Error())
				return
			}

			n.StartDiscovery(cfg.Topic)
			progressData.Set("Node started and discovery enabled")
			toggle.SetText("Turn Off")
		} else {
			progressData.Set("Stopping node...")
			n.Close()
			progressData.Set("Node stopped")
			toggle.SetText("Turn On")
		}
	})

	// Create a refresh button to manually refresh the views
	refreshButton := widget.NewButton("Refresh", refreshViews)

	// Layout the GUI elements in a border layout
	top := container.NewHBox(nodeNameWrap, statusWrap, nodeNameFieldWrap, nameEntryWrap, toggle, refreshButton)
	left := container.NewBorder(
		widget.NewLabel("Connected Peers"), // top
		nil,                                // bottom
		nil,                                // left
		nil,                                // right
		peerList,                           // center fills remaining space
	)
	appPanel := container.NewVBox()
	for _, registration := range apps.All() {
		registration := registration
		buttonLabel := "Run " + registration.App.Name
		if registration.Open != nil {
			buttonLabel = "Open " + registration.App.Name
		}
		appButton := widget.NewButton(buttonLabel, func() {
			if n == nil || !n.IsRunning() || selectedNode.AddrInfo.ID == "" {
				progressData.Set("Select a connected peer before running an app")
				return
			}
			if registration.Open != nil {
				registration.Open(n, selectedNode)
				return
			}

			peer := selectedNode
			progressData.Set("Running " + registration.App.Name + " for " + peer.Name + "...")
			go func() {
				ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
				defer cancel()

				reply, err := registration.Run(ctx, n, peer.AddrInfo.ID)
				if err != nil {
					progressData.Set(registration.App.Name + " failed: " + err.Error())
					return
				}
				progressData.Set(registration.App.Name + " reply from " + peer.Name + ": " + reply)
			}()
		})
		appButton.Disable()
		appButtons = append(appButtons, appButton)
		appPanel.Add(appButton)
	}
	right := container.NewBorder(
		widget.NewLabel("Apps (select a peer first)"),
		nil,
		nil,
		nil,
		appPanel,
	)
	mainCon := container.NewHSplit(left, right)
	mainCon.Offset = 0.30
	bottom := container.NewPadded(progressLabel)

	// Set the content of the window and run the application
	w.SetContent(container.NewBorder(top, bottom, nil, nil, mainCon))
	w.ShowAndRun()
}
