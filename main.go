package main

import (
	"flag"
	"fmt"

	"github.com/TilpDatLasse/HeisLab2025/HRA"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/network"
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
	flag.IntVar(&udpWVPort, "udpVWPort", 14700, "udp worldviews port")
	flag.IntVar(&peersPort, "peerPort", 16500, "online peers port")
	flag.Parse()

	SingElevChans := elev_algo.SingleElevatorChans{
		Drv_buttons:       make(chan elev.ButtonEvent),
		Drv_floors:        make(chan int),
		Drv_obstr:         make(chan bool),
		Drv_stop:          make(chan bool),
		Timer_channel:     make(chan bool),
		Single_elev_queue: make(chan [][2]bool),
	}

	WorldViewChans := worldview.WVChans{
		WorldViewTxChan: make(chan worldview.WorldView),
		WorldViewRxChan: make(chan worldview.WorldView),
	}

	ch_fromSync := make(chan map[string]worldview.InformationElev)
	ch_shouldSync := make(chan bool)

	ch_syncRequestsSingleElev := make(chan [][2]elev.ConfirmationState)

	go elev_algo.ElevMain(SingElevChans, ch_syncRequestsSingleElev, simPort)
	go network.NetworkMain(id, peersPort, WorldViewChans, udpWVPort)
	go HRA.HRAMain(SingElevChans.Single_elev_queue, ch_shouldSync, ch_fromSync, id)
	go worldview.SetElevatorStatus(WorldViewChans.WorldViewTxChan)
	go worldview.WorldViewMain(WorldViewChans.WorldViewRxChan, ch_syncRequestsSingleElev, ch_shouldSync, id)
	go syncing.SyncingMain(ch_shouldSync, ch_fromSync, ch_syncRequestsSingleElev)

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
// sjekke hva som egt skjer når den får to ordre i samme etasje (opp og ned), virker ikke som den venter 3 sek
// endre alt til camelcase(?)

//HUSKE Å SJEKKE UT HVORFOR HEISEN IKKE STOPPER I 3 SEK I ANDRE ETSASJE NÅR DEN GÅR FRA 1-2-3

// - 3 sekunder problemet
// - write og iterate problemet
// - hvordan fikser vi at en heis plutselig er stuck mellom to etasjer? vi må restarte den  -  her må vi få en av de andre heisene til å ta den
// - dra ut strømmen. Heis er fortsatt online? Hra kan fortsatt gi den ordre.
