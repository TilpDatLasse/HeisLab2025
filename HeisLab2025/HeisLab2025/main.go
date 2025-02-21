package main

import (

	/*
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
		"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
	*/
	"github.com/TilpDatLasse/HeisLab2025/elev_algo"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
)

func main() {
	go elev_algo.Elevator_hoved()
	go nettverk.Nettverk_hoved()
	for {
		select {}
	}
}
