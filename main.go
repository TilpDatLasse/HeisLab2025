package main

import (
	"flag"
	"fmt"
	"time"

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
	)

	SingElevChans := elev_algo.SingleElevatorChans{
		DrvButtons:      make(chan elev.ButtonEvent),
		DrvFloors:       make(chan int),
		DrvObstr:        make(chan bool),
		DrvStop:         make(chan bool),
		TimerChannel:    make(chan bool),
		SingleElevQueue: make(chan [][2]bool),
	}

	WorldViewChans := worldview.WVChans{
		WorldViewTxChan: make(chan worldview.WorldView),
		WorldViewRxChan: make(chan worldview.WorldView),
	}

	SyncChans := syncing.SyncChans{
		ShouldSync:              make(chan bool),
		InformationElevFromSync: make(chan map[string]worldview.InformationElev),
		SyncRequestSingleElev:   make(chan [][2]elev.ConfirmationState),
	}

	flag.StringVar(&id, "id", "one", "id of this peer")
	flag.StringVar(&simPort, "simPort", "15657", "simulation server port")
	flag.IntVar(&udpWVPort, "udpVWPort", 14700, "udp worldviews port")

	flag.Parse()

	go elev_algo.ElevMain(SingElevChans, SyncChans.SyncRequestSingleElev, simPort, id)

	// Sleep when initializing to make sure the elevator is ready
	time.Sleep(3 * time.Second)

	go network.NetworkMain(id, WorldViewChans, udpWVPort)
	go HRA.HRAMain(SingElevChans.SingleElevQueue, SyncChans.ShouldSync, SyncChans.InformationElevFromSync, id)
	go worldview.SetElevatorStatus(WorldViewChans.WorldViewTxChan)
	go worldview.WorldViewMain(WorldViewChans.WorldViewRxChan, SyncChans.SyncRequestSingleElev, SyncChans.ShouldSync, id)
	go syncing.SyncingMain(SyncChans)

	select {}

}
