package sync

import (
	"fmt"
	"reflect"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
)

// Må sende denne heisen sitt verdensbilde
// Må motta verdensbildet til de andre heisene
// Må synkronisere med cyclic counter slik at alle heiser er synkronisert

//må oppdatere elevator-variabalen i fsm så den vet om en ordre blir tatt av noen andre

//Forslag til hvordan synkingen kan løses:
//Legge inn en til confirmationstat-variabel ("Locked") som brukes for å vise andre heiser at man er klar
// til å sende verdensbildet sitt til hra. når alle er enige om å sende (samme) verdensbilde må verdensbiledene
// låses og kan ikke endres før alle bekrefter at de har sendt til hra elns.
// Kan starte med at hra-en sender en request om å få status, og da henter heisen info fra sin heis og låser,
// før den broadcaster at den er klar til å sende til hra. når andre heiser mottar dette må de gjøre det samme,
// til alle er klar til å sende samme verdensbilde til hra og gjør dette.

// synker wv før infomap sendes til hra
// sammenligner alle hallrequest-listene som ligger i hra og synkroniserer/oppdaterer dersom confirmationstate
// tillater det.
// broadcaster og sammenlikner også hele worldView til hver peer, så alle sender akkurat samme info til HRA

func Sync() {
	for{
		if AllWorldViewsEqual(nettverk.WorldViewMap) {
			ch_shouldSync <- false
			break
			//fmt.Println("All worldviews are equal")
		} else {
			//fmt.Println("Worldviews are not equal")
		}
	}
}

func CompareAndUpdateWV(ch_toSync chan map[string]nettverk.InformationElev, ch_fromSync chan map[string]nettverk.InformationElev) {
	infoMap := <-ch_toSync
	for allSynced(infoMap) {
		if len(infoMap) == 0 {
			//panic("Infomap er tomt!")
		} else {
			for i := 0; i < elev.N_FLOORS; i++ {
				for j := 0; j < elev.N_BUTTONS-1; j++ {
					listOfElev := make([]elev.ConfirmationState, len(infoMap))
					k := 0
					for _, elev := range infoMap {
						listOfElev[k] = elev.HallRequests[i][j]
						k++
					}
					update := cyclicUpdate(listOfElev)

					for _, elev := range infoMap {
						elev.HallRequests[i][j] = update
					}
				}
			}

		}
		ch_fromSync <- infoMap
	}
}

// hjelpefunksjon for CompareAndUpdateWV
func cyclicUpdate(list []elev.ConfirmationState) elev.ConfirmationState {
	isPresent := map[elev.ConfirmationState]bool{} // map som lagrer om hver confimationstate(0,1,2) er tilstede
	for _, v := range list {
		isPresent[v] = true
	}
	switch {
	case isPresent[0] && isPresent[1] && isPresent[2]:
		panic("Confirmationstates 0,1,2 at the same time :(")
	case !isPresent[0]: // alle har 1 eller 2
		fmt.Println("Order registrerd on all peers, Confirmed!")
		return 2
	case isPresent[2] && isPresent[0]: // alle har 0 eller 2 (noen har utført ordren)
		return 0
	case isPresent[0] && isPresent[1]: // alle har 0 eller 1 (noen har fått en ny ordre)
		return 1
	}
	return 0 //default
}

// func allSynced(wvMap map[string]nettverk.WorldView) bool {
// 	var firstElev nettverk.InformationElev
// 	isFirst := true

// 	for _, elev := range wvMap {
// 		if isFirst {
// 			firstElev.HallRequests = elev.HallRequests
// 			isFirst = false
// 		} else {
// 			for i := 0; i < len(m); i++ {
// 				firstPair := firstElev.HallRequests[i]
// 				for _, s := range elev.HallRequests[1:] {
// 					if s != firstPair {
// 						return false
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return true
// }

func AllWorldViewsEqual(worldViewMap map[string]nettverk.WorldView) bool {
	var reference *nettverk.WorldView

	// Gå gjennom alle verdiene i mappet
	for _, worldView := range worldViewMap {
		if reference == nil {
			// Sett første `WorldView` som referanse
			reference = &worldView
			continue
		}

		// Sammenlign hele `InfoMap`
		if !reflect.DeepEqual(reference.InfoMap, worldView.InfoMap) {
			return false
		}
	}

	return true
}

// returnerer true hvis alle heiser holder samme liste med hallrequests
// må endre så den også sammenlikner andre egenskaper, og sjekker om alle er låst
// func allSynced(m map[string]nettverk.InformationElev) bool {
// 	var firstElev nettverk.InformationElev
// 	isFirst := true

// 	for _, elev := range m {
// 		if isFirst {
// 			firstElev.HallRequests = elev.HallRequests
// 			isFirst = false
// 		} else {
// 			for i := 0; i < len(m); i++ {
// 				firstPair := firstElev.HallRequests[i]
// 				for _, s := range elev.HallRequests[1:] {
// 					if s != firstPair {
// 						return false
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return true
// }
