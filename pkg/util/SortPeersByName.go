package util

import (
	"fgov/network/pkg/node"
	"sort"
)

func SortPeersByName(peers []node.PeerRecord) []node.PeerRecord {
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Name < peers[j].Name
	})

	return peers
}
