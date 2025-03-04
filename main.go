package main

import (

	/*
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
	*/
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"

	//"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
	//"github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/HRA"
)

func main() {

	ElevatorStatusChannel := make(chan elev.Elevator)
	HRAOut := make(chan map[string][][2]bool)
	ch_single_elevator_qeue := make(chan [][2]bool)

	ch_HRAInputTx := make(chan nettverk.InformationElev)
	ch_HRAInputRx := make(chan nettverk.InformationElev)
	//peerToHRA := make(chan peers.PeerUpdate)
	//change_ch := make(chan bool)

	go elev_algo.Elevator_hoved(ch_single_elevator_qeue)
	go nettverk.Nettverk_hoved(ch_HRAInputRx)
	go HRA.HRAMain(HRAOut)
	go nettverk.SetElevatorStatus(ch_HRAInputTx)
	go nettverk.RecieveElevatorStatus(ch_HRAInputRx)
	go nettverk.BroadcastElevatorStatus(ch_HRAInputTx, ElevatorStatusChannel)
	//go fsm.FetchElevatorStatus()
	//go bcast.Receiver(17000, buttonRx)
	for {
		select {}
	}
}
