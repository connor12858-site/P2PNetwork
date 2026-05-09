package main

import (
	"fmt"
	"os"

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
    w.SetFixedSize(true)

    var n *node.Node
    var toggle *widget.Button

    status := widget.NewLabel("Stopped")
    peerIDLabel := widget.NewLabel("Peer ID: (not started)")

    peersBox := container.NewVBox()
    appsBox := container.NewVBox()

    var peersScroll *container.Scroll
    var appsScroll *container.Scroll

    refreshViews := func() {
        peersBox.Objects = nil
        appsBox.Objects = nil

        if n != nil && n.IsRunning() {
            connectedPeers := n.GetPeersSnapshot()

            for _, p := range connectedPeers {
                peersBox.Add(widget.NewLabel(fmt.Sprintf("%s", p.ID)))
            }

            // Auto-expand peers scroll height based on number of peers
            if peersScroll != nil {
                desiredHeight := float32(len(connectedPeers)*24 + 10)
                maxHeight := float32(360)
                if desiredHeight > maxHeight {
                    desiredHeight = maxHeight
                }
                peersScroll.SetMinSize(fyne.NewSize(320, desiredHeight))
            }

            status.SetText(fmt.Sprintf("Running - %d peers", len(connectedPeers)))
            peerIDLabel.SetText("Peer ID: " + n.Host.ID().String())
        } else {
            status.SetText("Stopped")
            peerIDLabel.SetText("Peer ID: (not started)")
            if peersScroll != nil {
                peersScroll.SetMinSize(fyne.NewSize(320, 80))
            }
        }

        peersBox.Refresh()
        appsBox.Refresh()
    }

    toggle = widget.NewButton("Turn On", func() {
        if n == nil || !n.IsRunning() {
            bootstrapPeers := readBootstrap()
            var err error
            n, err = node.NewNode(0, bootstrapPeers)
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

    peersScroll = container.NewVScroll(peersBox)
    appsScroll = container.NewVScroll(appsBox)

    left := container.NewVBox(peerIDLabel, status, toggle, refreshButton, widget.NewLabel("Connected Peers"), peersScroll)
    right := container.NewVBox(widget.NewLabel("Apps"), appsScroll)

    w.SetContent(container.NewBorder(nil, nil, left, nil, right))
    w.ShowAndRun()
}
