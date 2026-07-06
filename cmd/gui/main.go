package main

import (
	"fmt"
	"os"
	"strings"

	"fgov/network/pkg/node"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Global variables
var BOOTSTRAP_ADDRESS = ""
var BOOTSTRAP_PEERS = make([]string, 0, 1)
var TOPIC = ""
var PORT = 0

// config reads the configuration from config.yaml and sets the global variables.
func config() {
    // Check for the yaml file first
    if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
        os.Exit(1)
    }

    // Gather the data
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        fmt.Println("Error reading config.yaml:", err)
        os.Exit(1)
    }

    // Save the data
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "port:") {
            fmt.Sscanf(line, "port: %d", &PORT)
        } else if strings.HasPrefix(line, "topic:") {   
            fmt.Sscanf(line, "topic: %s", &TOPIC)
        } else if strings.HasPrefix(line, "bootstrap:") {
            fmt.Sscanf(line, "bootstrap: %s", &BOOTSTRAP_ADDRESS)
        }
    }
}

// Reads bootstrap peers from bs-nodes file.
func get_bootstrap() {
    data, err := os.ReadFile("bs-nodes")
    if err == nil && len(strings.TrimSpace(string(data))) > 0 {
        bootstrap := strings.TrimSpace(string(data))
        fmt.Println("Bootstrap multiaddr from file:", bootstrap)
        BOOTSTRAP_PEERS = append(BOOTSTRAP_PEERS, bootstrap)
    }
}

// main launches the GUI application.
func main() {
    a := app.New()
    w := a.NewWindow("P2P Network")
    w.Resize(fyne.NewSize(720, 480))

    var n *node.Node
    var toggle *widget.Button

    status := widget.NewLabel("Stopped")
    nodeNameLabel := widget.NewLabel("Name: (not started)")
    
    nameEntry := widget.NewEntry()
    nameEntry.SetPlaceHolder("Enter node name...")

    peersBox := container.NewVBox()
    appsBox := container.NewVBox()

    refreshViews := func() {
        peersBox.Objects = nil

        if n != nil && n.IsRunning() {
            connectedPeers := n.GetPeersSnapshot()

            for _, p := range connectedPeers {
                if !strings.Contains(p.Name, "-bootstrap") {
                    peersBox.Add(widget.NewLabel(p.Name))
                }
            }

            status.SetText(fmt.Sprintf("Running - %d peers", len(connectedPeers)))
            nodeNameLabel.SetText("Name: " + n.Name)
        } else {
            status.SetText("Stopped")
            nodeNameLabel.SetText("Name: (not started)")
        }

        peersBox.Refresh()
    }

    toggle = widget.NewButton("Turn On", func() {
        if n == nil || !n.IsRunning() {
            var err error
            n, err = node.NewNode(PORT, BOOTSTRAP_PEERS, strings.ReplaceAll(nameEntry.Text, "-bootstrap", ""))
            if err != nil {
                status.SetText("Start error: " + err.Error())
                return
            }

            n.StartDiscovery(TOPIC)
            toggle.SetText("Turn Off")
        } else {
            n.Close()
            toggle.SetText("Turn On")
        }

        refreshViews()
    })

    refreshButton := widget.NewButton("Refresh", refreshViews)

    peersScroll := container.NewVScroll(peersBox)
    appsScroll := container.NewVScroll(appsBox)

    left := container.NewVBox(nodeNameLabel, status, widget.NewLabel("Node Name:"), nameEntry, toggle, refreshButton, widget.NewLabel("Connected Peers"), peersScroll)
    right := container.NewVBox(widget.NewLabel("Apps"), appsScroll)

    w.SetContent(container.NewBorder(nil, nil, left, nil, right))
    w.ShowAndRun()
}
