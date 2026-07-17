package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fgov/network/pkg/node"
	"fgov/network/pkg/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

	status := widget.NewLabel("Stopped")
	nodeNameLabel := widget.NewLabel("Name: (not started)")
	progressLabel := widget.NewLabel("Ready")
	progressLabel.Wrapping = fyne.TextWrapWord

	nameEntry := widget.NewEntry()
	nameEntry.SetText(cfg.Name) // Set the default name from config
	nameEntryWrap := container.NewGridWrap(fyne.NewSize(240, nameEntry.MinSize().Height), nameEntry)

	peersBox := container.NewVBox()
	appsBox := container.NewVBox()

	refreshViews := func() {
		peersBox.Objects = nil

		if n != nil && n.IsRunning() {
			connectedPeers := n.GetPeersSnapshot()
			connectedPeers = util.SortPeersByName(connectedPeers) // Sort peers by name

			for _, p := range connectedPeers {
				if !strings.Contains(p.Name, "-bootstrap") {
					peersBox.Add(widget.NewLabel(p.Name))
				}
			}

			status.SetText(fmt.Sprintf("Running - %d peers", len(connectedPeers)))
			nodeNameLabel.SetText("Name: " + n.Name)
			nameEntry.Disable() // Disable the name entry when the node is running
		} else {
			status.SetText("Stopped")
			nodeNameLabel.SetText("Name: (not started)")
			nameEntry.Enable() // Enable the name entry when the node is stopped
		}

		peersBox.Refresh()
	}

	// Periodically refresh the views every 5 seconds
	go func() {
		for {
			refreshViews()
			time.Sleep(5 * time.Second)
		}
	}()

	toggle = widget.NewButton("Turn On", func() {
		if n == nil || !n.IsRunning() {
			var err error
			progressLabel.SetText("Starting node...")
			n, err = node.NewNode(cfg.Port, BOOTSTRAP_PEERS, strings.ReplaceAll(nameEntry.Text, "-bootstrap", ""))
			if err != nil {
				status.SetText("Start error: " + err.Error())
				progressLabel.SetText("Start failed: " + err.Error())
				return
			}

			n.StartDiscovery(cfg.Topic)
			progressLabel.SetText("Node started and discovery enabled")
			toggle.SetText("Turn Off")
		} else {
			progressLabel.SetText("Stopping node...")
			n.Close()
			progressLabel.SetText("Node stopped")
			toggle.SetText("Turn On")
		}

		refreshViews()
	})

	refreshButton := widget.NewButton("Refresh", refreshViews)

	peersScroll := container.NewVScroll(peersBox)
	peersScroll.SetMinSize(fyne.NewSize(200, 300))
	appsScroll := container.NewVScroll(appsBox)

	top := container.NewHBox(nodeNameLabel, status, widget.NewLabel("Node Name:"), nameEntryWrap, toggle, refreshButton)
	left := container.NewVBox(widget.NewLabel("Connected Peers"), peersScroll)
	right := container.NewVBox(widget.NewLabel("Apps"), appsScroll)
	main_con := container.NewHSplit(left, right)
	bottom := container.NewPadded(progressLabel)

	w.SetContent(container.NewBorder(top, bottom, nil, nil, main_con))
	w.ShowAndRun()
}
