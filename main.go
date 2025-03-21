package main

import (
	"flag"
	"fmt"

	//"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
	"github.com/TilpDatLasse/HeisLab2025/syncing"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

func main() {
	fmt.Println("Starting, please wait...")

	var (
		id        string
		simPort   string
		udpWVPort int
		peersPort int
	)

	flag.StringVar(&id, "id", "one", "id of this peer")
	flag.StringVar(&simPort, "simPort", "15657", "simulation server port")
	flag.IntVar(&udpWVPort, "udpVWPort", 14700, "sudp worldviews port")
	flag.IntVar(&peersPort, "peerPort", 16500, "onlie peers port")
	flag.Parse()

	SingElevChans := elev_algo.SingleElevatorChans{
		Drv_buttons:       make(chan elev.ButtonEvent),
		Drv_floors:        make(chan int),
		Drv_obstr:         make(chan bool),
		Drv_stop:          make(chan bool),
		Timer_channel:     make(chan bool),
		Single_elev_queue: make(chan [][2]bool),
	}

	//ch_toSync := make(chan map[string]nettverk.InformationElev)   //sender infomap
	ch_fromSync := make(chan map[string]worldview.InformationElev) //sender infomap
	ch_shouldSync := make(chan bool)
	//ch_allSynced := make(chan bool)
	ch_WVRx := make(chan worldview.WorldView)
	ch_WVTx := make(chan worldview.WorldView)
	ch_syncRequestsSingleElev := make(chan [][2]elev.ConfirmationState)

	go elev_algo.Elev_main(SingElevChans, ch_syncRequestsSingleElev, simPort)
	go nettverk.Nettverk_hoved(ch_WVRx, id, peersPort)
	//go sync.Sync()

	go HRA.HRAMain(SingElevChans.Single_elev_queue, ch_shouldSync, ch_fromSync, id)
	go worldview.SetElevatorStatus(ch_WVTx)
	go nettverk.RecieveWV(ch_WVRx, udpWVPort)
	go nettverk.BroadcastWV(ch_WVTx, udpWVPort)
	go worldview.WorldViewFunc(ch_WVRx, ch_syncRequestsSingleElev, ch_shouldSync, id)
	go syncing.Syncing(ch_shouldSync, ch_fromSync)

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

//TODO:
// Sende på channel fra elevalgo til hra når noe skjer, så hra kan sende syncrequest (husk å bruke select når vi skriver til channel)
// Dele opp i flere moduler (syncing, WorldView, etc.)
// Rydde bort unnødvendig quick-fixes
// Robusthet mtp packetloss
// Locked-variabelen er nok ikke helt robust (?)
