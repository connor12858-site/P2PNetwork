package text

import (
	"context"
	"strings"
	"time"

	"fgov/network/pkg/node"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// Open shows the current conversation with peerRecord and allows new messages.
func Open(n *node.Node, peerRecord node.PeerRecord) {
	w := fyne.CurrentApp().NewWindow("Messages — " + peerRecord.Name)
	w.Resize(fyne.NewSize(440, 520))

	messageData := binding.NewString()
	messageList := widget.NewLabelWithData(messageData)
	messageList.Wrapping = fyne.TextWrapWord
	refresh := func() {
		messageData.Set(render(Conversation(n, peerRecord.AddrInfo.ID)))
	}
	refresh()

	// entry is the text entry for new messages, and status shows the current status of sending.
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Write a message...")
	entry.Wrapping = fyne.TextWrapWord
	status := widget.NewLabel("")

	// sendMessage sends the message in the entry to the peer and clears the entry.
	sendMessage := func() {
		text := strings.TrimSpace(entry.Text)
		if text == "" {
			return
		}
		entry.SetText("")
		status.SetText("Sending...")

		go func() {
			ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
			defer cancel()
			if err := Send(ctx, n, peerRecord.AddrInfo.ID, text); err != nil {
				status.SetText("Send failed: " + err.Error())
				return
			}
			status.SetText("")
			refresh()
		}()
	}

	send := widget.NewButton("Send", sendMessage)
	entry.OnSubmitted = func(_ string) {
		sendMessage()
	}

	stopRefresh := make(chan struct{})
	w.SetOnClosed(func() { close(stopRefresh) })
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				refresh()
			case <-stopRefresh:
				return
			}
		}
	}()

	w.SetContent(container.NewBorder(
		nil,
		container.NewVBox(status, container.NewBorder(nil, nil, nil, send, entry)),
		nil,
		nil,
		container.NewVScroll(messageList),
	))
	w.Show()
}
