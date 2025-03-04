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

	SingElevChans := elev_algo.SingleElevatorChans{
		Drv_buttons:       make(chan elev.ButtonEvent),
		Drv_floors:        make(chan int),
		Drv_obstr:         make(chan bool),
		Drv_stop:          make(chan bool),
		Timer_channel:     make(chan bool),
		Single_elev_queue: make(chan [][2]bool),
	}

	ch_ElevatorStatus := make(chan elev.Elevator)
	ch_HRAOut := make(chan map[string][][2]bool)
	ch_HRAInputTx := make(chan nettverk.InformationElev)
	ch_HRAInputRx := make(chan nettverk.InformationElev)

	go elev_algo.Elevator_hoved(SingElevChans)
	go nettverk.Nettverk_hoved(ch_HRAInputRx)
	go HRA.HRAMain(ch_HRAOut)
	go nettverk.SetElevatorStatus(ch_HRAInputTx)
	go nettverk.RecieveElevatorStatus(ch_HRAInputRx)
	go nettverk.BroadcastElevatorStatus(ch_HRAInputTx, ch_ElevatorStatus)
	go nettverk.FromHRA(ch_HRAOut, SingElevChans.Single_elev_queue)
	//go fsm.FetchElevatorStatus()
	//go bcast.Receiver(17000, buttonRx)
	for {
		select {}
	}
}
