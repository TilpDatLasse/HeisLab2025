package peers

import (
	"fmt"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
	Id    string
}

const (
	interval = 15 * time.Millisecond
	timeout  = 5000 * time.Millisecond // 5 seconds to determine if peer is lost or not (packetloss crucial)
)

var (
	PeerToUpdate   = make(chan PeerUpdate)
	PeerFromUpdate = make(chan PeerUpdate)
)

// Update peers that are online
func UpdatePeers() {
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	for {
		updated := false
		p = <-PeerToUpdate
		id := p.Id

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}
		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}
		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))
			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}
			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			PeerFromUpdate <- p
		}
	}
}

func PrintPeerUpdate(pUpdate PeerUpdate) {

	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", pUpdate.Peers)
	fmt.Printf("  New:      %q\n", pUpdate.New)
	fmt.Printf("  Lost:     %q\n", pUpdate.Lost)

}
