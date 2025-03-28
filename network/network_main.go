package network

import (
	b "github.com/TilpDatLasse/HeisLab2025/network/bcast"
	p "github.com/TilpDatLasse/HeisLab2025/network/peers"
	wv "github.com/TilpDatLasse/HeisLab2025/worldview"
)

func BroadcastWV(WVTxChan chan wv.WorldView, udpWVPort int) {
	b.Transmitter(udpWVPort, WVTxChan)
}

func RecieveWV(WVRxChan chan wv.WorldView, udpWVPort int) {
	b.Receiver(udpWVPort, WVRxChan)
}

func NetworkMain(id string, wvChans wv.WVChans, udpWVPort int) {

	go RecieveWV(wvChans.WorldViewRxChan, udpWVPort)
	go BroadcastWV(wvChans.WorldViewTxChan, udpWVPort)
	go p.UpdatePeers()

	for {
		pUpdate := <-p.PeerFromUpdate

		p.PrintPeerUpdate(pUpdate)

		// Functionalities for updating the world view and deleting lost peers
		for i := 0; i < len(pUpdate.Lost); i++ {
			lostpeer := pUpdate.Lost[i]
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
