package nettverk

import (
	"fmt"

	b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

func BroadcastWV(ch_WVTx chan worldview.WorldView, udpWVPort int) {
	b.Transmitter(udpWVPort, ch_WVTx)
}

func RecieveWV(ch_WVRx chan worldview.WorldView, udpWVPort int) {
	b.Receiver(udpWVPort, ch_WVRx)
}

func Nettverk_hoved(ch_WVRx chan worldview.WorldView, id string, peerPort int) {

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(peerPort, id, peerTxEnable)
	go peers.Receiver(peerPort, peerUpdateCh)

	for {
		p := <-peerUpdateCh
		fmt.Printf("Peer update:\n")
		fmt.Printf("  Peers:    %q\n", p.Peers)
		fmt.Printf("  New:      %q\n", p.New)
		fmt.Printf("  Lost:     %q\n", p.Lost)
		if len(p.Lost) != 0 {
			for i := 0; i < len(p.Lost); i++ {
				lostpeer := p.Lost[i]
				delete(worldview.InfoMap, lostpeer)
				delete(worldview.WorldViewMap, lostpeer)
			}
		}
	}
}
