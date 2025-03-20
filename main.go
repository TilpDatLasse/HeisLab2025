package main

import (
	"flag"
	"fmt"

	//"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
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
	//ch_toSync := make(chan map[string]nettverk.InformationElev)   //sender infomap
	ch_fromSync := make(chan map[string]nettverk.InformationElev) //sender infomap
	ch_shouldSync := make(chan bool)
	//ch_allSynced := make(chan bool)
	ch_WVRx := make(chan nettverk.WorldView)
	ch_WVTx := make(chan nettverk.WorldView)
	ch_syncRequestsSingleElev := make(chan [][2]elev.ConfirmationState)

	go elev_algo.Elev_main(SingElevChans, ch_syncRequestsSingleElev, simPort)
	go nettverk.Nettverk_hoved(ch_HRAInputRx, ch_WVRx, ch_shouldSync, ch_fromSync, ch_syncRequestsSingleElev, id)
	//go sync.Sync()

	go HRA.HRAMain(ch_HRAOut, ch_shouldSync, ch_fromSync)
	go nettverk.SetElevatorStatus(ch_HRAInputTx, ch_WVTx)
	go nettverk.RecieveWV(ch_WVRx)
	go nettverk.BroadcastWV(ch_WVTx)
	go nettverk.FromHRA(ch_HRAOut, SingElevChans.Single_elev_queue)

	select {}

}

//kommentarer fra studass:
//Netteverksmodul kan brytes ned, moduler burde generelt innholde ting som er nært deres kjerneoppgave
//Mainfilen er fint strukturert
//cyclic-counter er viktig, kan implementeres i den opprinnelige hallrequest-listen
//Worldview er veldig viktig og burde nok være egen modul
//channels som bare går inn i kun én funksjon er sannsynligvis overflødige.

//viser seg å være lurt å legge til en select når vi skriver til kanal, så blokkerer den ikke selv om den er full (men mister selvfølgelig info)

//må huske å endre så hra sycer når det skjer noe, ikke bare regelmessig
//burde kalle sync-modulen noe annet, mutex-modulen heter også sync

//virker egt som om lys funker som de skal nå
