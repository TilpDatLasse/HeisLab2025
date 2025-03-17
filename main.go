package main

import (
	"flag"
	"fmt"

	//"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
	"github.com/TilpDatLasse/HeisLab2025/sync"
)

func main() {
	fmt.Println("Starting, please wait...")

	var (
		id      string
		simPort string
	)

	flag.StringVar(&id, "id", "one", "id of this peer")
	flag.StringVar(&simPort, "simPort", "15657", "simulation server port")
	flag.Parse()

	SingElevChans := elev_algo.SingleElevatorChans{
		Drv_buttons:       make(chan elev.ButtonEvent),
		Drv_floors:        make(chan int),
		Drv_obstr:         make(chan bool),
		Drv_stop:          make(chan bool),
		Timer_channel:     make(chan bool),
		Single_elev_queue: make(chan [][2]bool),
	}

	ch_HRAOut := make(chan map[string][][2]bool)
	ch_HRAInputTx := make(chan nettverk.InformationElev)
	ch_HRAInputRx := make(chan nettverk.InformationElev)
	ch_toSync := make(chan map[string]nettverk.InformationElev)   //sender infomap
	ch_fromSync := make(chan map[string]nettverk.InformationElev) //sender infomap

	go elev_algo.Elev_main(SingElevChans, simPort)
	go nettverk.Nettverk_hoved(ch_HRAInputRx, id)
	go HRA.HRAMain(ch_HRAOut, ch_toSync, ch_fromSync)
	go nettverk.SetElevatorStatus(ch_HRAInputTx)
	go nettverk.RecieveElevatorStatus(ch_HRAInputRx)
	go nettverk.BroadcastElevatorStatus(ch_HRAInputTx)
	go nettverk.FromHRA(ch_HRAOut, SingElevChans.Single_elev_queue)
	go sync.CompareAndUpdateWV(ch_toSync, ch_fromSync)

	select {}

}

//kommentarer fra studass:
//Netteverksmodul kan brytes ned, moduler burde generelt innholde ting som er nært deres kjerneoppgave
//Mainfilen er fint strukturert
//cyclic-counter er viktig, kan implementeres i den opprinnelige hallrequest-listen
//Worldview er veldig viktig og burde nok være egen modul
//channels som bare går inn i kun én funksjon er sannsynligvis overflødige.
