package main

import (

	/*
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
	*/
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"

	//"github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/HRA"
)

func main() {

	ElevatorStatusChannel := make(chan elev.Elevator)

	buttonTx := make(chan elev.ButtonEvent)
	buttonRx := make(chan elev.ButtonEvent)

	buttonRx1 := make(chan elev.ButtonEvent)

	ch_HRAInputTx := make(chan HRA.HRAInput)
	ch_HRAInputRx := make(chan HRA.HRAInput)
	change_ch := make(chan bool)
	go fsm.FetchElevatorStatus(ElevatorStatusChannel)
	go elev_algo.Elevator_hoved(buttonTx, buttonRx1, change_ch)
	go nettverk.Nettverk_hoved(buttonTx, buttonRx, buttonRx1)
	go HRA.HRAMain(ch_HRAInputTx, ch_HRAInputRx, change_ch, ElevatorStatusChannel)
	//go bcast.Receiver(17000, buttonRx)
	for {
		select {}
	}
}
