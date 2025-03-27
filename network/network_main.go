package network

import (
	b "github.com/TilpDatLasse/HeisLab2025/network/bcast"
	p "github.com/TilpDatLasse/HeisLab2025/network/peers"
	wv "github.com/TilpDatLasse/HeisLab2025/worldview"
)

func BroadcastWV(ch_WVTx chan wv.WorldView, udpWVPort int) {
	b.Transmitter(udpWVPort, ch_WVTx)
}

func RecieveWV(ch_WVRx chan wv.WorldView, udpWVPort int) {
	b.Receiver(udpWVPort, ch_WVRx)
}

func NetworkMain(id string, wvChans wv.WVChans, udpWVPort int) {

	go RecieveWV(wvChans.WorldViewRxChan, udpWVPort)
	go BroadcastWV(wvChans.WorldViewTxChan, udpWVPort)
	go p.UpdatePeers()

	for {
		Pupdate := <-p.PeerFromUpdate

		p.PrintPeerUpdate(Pupdate)

		for i := 0; i < len(Pupdate.Lost); i++ {
			lostpeer := Pupdate.Lost[i]
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
