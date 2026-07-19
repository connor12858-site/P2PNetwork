// Package text implements the first version of direct peer-to-peer messaging.
package text

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"fgov/network/pkg/node"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ID       = "text"
	Name     = "Messages"
	Protocol = protocol.ID("/p2pnetwork/text/1.0.0")
)

// Message is one plain-text message exchanged by two peers.
type Message struct {
	From string    `json:"from"`
	Text string    `json:"text"`
	Sent time.Time `json:"sent"`
}

var conversations = struct {
	sync.RWMutex
	items map[string][]Message
}{items: make(map[string][]Message)}

// Register exposes the Messages protocol on a node.
func Register(n *node.Node) error {
	return n.RegisterApp(node.App{ID: ID, Name: Name, Protocol: Protocol}, handleStream(n))
}

// Send delivers a message to peerID and records it locally after the stream is accepted.
func Send(ctx context.Context, n *node.Node, peerID peer.ID, text string) error {
	message := Message{From: n.Name, Text: strings.TrimSpace(text), Sent: time.Now()}
	if message.Text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	stream, err := n.OpenAppStream(ctx, peerID, ID)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := json.NewEncoder(stream).Encode(message); err != nil {
		return err
	}
	appendMessage(n, peerID, message)
	return nil
}

// Conversation returns a copy of the messages between this node and peerID.
func Conversation(n *node.Node, peerID peer.ID) []Message {
	conversations.RLock()
	defer conversations.RUnlock()

	messages := conversations.items[conversationKey(n, peerID)]
	return append([]Message(nil), messages...)
}

func handleStream(n *node.Node) network.StreamHandler {
	return func(stream network.Stream) {
		defer stream.Close()

		var message Message
		if err := json.NewDecoder(bufio.NewReader(stream)).Decode(&message); err != nil {
			fmt.Printf("Text app read failed: %v\n", err)
			return
		}
		message.Text = strings.TrimSpace(message.Text)
		if message.Text == "" {
			return
		}
		if message.Sent.IsZero() {
			message.Sent = time.Now()
		}
		appendMessage(n, stream.Conn().RemotePeer(), message)
	}
}

func appendMessage(n *node.Node, peerID peer.ID, message Message) {
	conversations.Lock()
	defer conversations.Unlock()
	key := conversationKey(n, peerID)
	conversations.items[key] = append(conversations.items[key], message)
}

func conversationKey(n *node.Node, peerID peer.ID) string {
	return n.Host.ID().String() + ":" + peerID.String()
}

func render(messages []Message) string {
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].Sent.Before(messages[j].Sent) })
	if len(messages) == 0 {
		return "No messages yet."
	}

	lines := make([]string, 0, len(messages))
	for _, message := range messages {
		lines = append(lines, fmt.Sprintf("%s  %s\n%s", message.Sent.Format("15:04"), message.From, message.Text))
	}
	return strings.Join(lines, "\n\n")
}
