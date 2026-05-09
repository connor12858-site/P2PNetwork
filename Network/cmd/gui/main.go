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

func readBootstrap() []string {
    data, err := os.ReadFile("bs-nodes")
    if err != nil {
        return nil
    }

    bootstrap := string(data)
    if bootstrap == "" {
        return nil
    }

    return []string{bootstrap}
}

func main() {
    a := app.New()
    w := a.NewWindow("FGov P2P Network")
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
            bootstrapPeers := readBootstrap()

            var err error
            n, err = node.NewNode(0, bootstrapPeers, strings.ReplaceAll(nameEntry.Text, "-bootstrap", ""))
            if err != nil {
                status.SetText("Start error: " + err.Error())
                return
            }

            n.StartDiscovery("fgov-network")
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
