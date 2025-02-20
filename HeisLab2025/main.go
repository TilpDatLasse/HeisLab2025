package main

import (

	/*
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network_go/network/bcast"
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network_go/network/peers"
	*/
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network_go"
)

func main() {
	go elev_algo.Elevator_hoved()
	go network_go.Nettverk_hoved()
	for {
		select {}
	}
}
