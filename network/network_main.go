package network

import (
	"fmt"

	b "github.com/TilpDatLasse/HeisLab2025/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/network/peers"
	wv "github.com/TilpDatLasse/HeisLab2025/worldview"
)

func BroadcastWV(ch_WVTx chan wv.WorldView, udpWVPort int) {
	b.Transmitter(udpWVPort, ch_WVTx)
}

func RecieveWV(ch_WVRx chan wv.WorldView, udpWVPort int) {
	b.Receiver(udpWVPort, ch_WVRx)
}

func NetworkMain(id string, peerPort int, wvChans wv.WVChans, udpWVPort int) {

	go RecieveWV(wvChans.WorldViewRxChan, udpWVPort)
	go BroadcastWV(wvChans.WorldViewTxChan, udpWVPort)
	go peers.UpdatePeers()

	for {
		p := <-peers.PeerFromUpdate
		fmt.Printf("Peer update:\n")
		fmt.Printf("  Peers:    %q\n", p.Peers)
		fmt.Printf("  New:      %q\n", p.New)
		fmt.Printf("  Lost:     %q\n", p.Lost)

		for i := 0; i < len(p.Lost); i++ {
			lostpeer := p.Lost[i]
			if lostpeer != id {
				wv.InfoMapMutex.Lock()
				delete(wv.InfoMap, lostpeer)
				wv.MyWorldView.InfoMap = wv.InfoMap
				wv.InfoMapMutex.Unlock()
				wv.WVMapMutex.Lock()
				delete(wv.WorldViewMap, lostpeer)
				wv.WorldViewMap[id] = wv.MyWorldView
				wv.WVMapMutex.Unlock()
			}
		}
	}
}
