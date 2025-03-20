package sync

import (
	"fmt"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
)

// Må sende denne heisen sitt verdensbilde
// Må motta verdensbildet til de andre heisene
// Må synkronisere med cyclic counter slik at alle heiser er synkronisert

//må oppdatere elevator-variabalen i fsm så den vet om en ordre blir tatt av noen andre

//Forslag til hvordan synkingen kan løses:
//Legge inn en til confirmationstat-variabel som brukes for å vise andre heiser at man er klar
// til å sende verdensbildet sitt til hra. når alle er enige om å sende (samme) verdensbilde må verdensbiledene
// låses og kan ikke endres før alle bekrefter at de har sendt til hra elns.
// Kan starte med at hra-en sender en request om å få status, og da henter heisen info fra sin heis og låser,
// før den broadcaster at den er klar til å sende til hra. når andre heiser mottar dette må de gjøre det samme,
// til alle er klar til å sende samme verdensbilde til hra og gjør dette.

// Hvis !allSynced etter at vi har kjørt compareAndUpdateWV én gang, må annen info for en av heisene være ulik,
// og vi må broadcaste på nytt for å få synket
// evt: alle broadcaster en gang etter de har blitt låst, da skal vel alle ende opp med samme?

// synker wv før infomap sendes til hra
// sammenligner alle hallrequest-listene som ligger i hra og synkroniserer/oppdaterer dersom confirmationstate
// tillater det.

func CompareAndUpdateWV(ch_toSync chan map[string]nettverk.InformationElev, ch_fromSync chan map[string]nettverk.InformationElev) {
	allSynced := false
	for !allSynced {
		infoMap := <-ch_toSync
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
		panic("Confirmationstates 0,1,2 at the same time :(") // denne linjen går ikke med tre heiser om man trykker på knappen som akkurat er i overgang fra 2 til 0 etter å ha blitt tatt. Da får man 0,1 og 2
	case !isPresent[0]: // alle har 1 eller 2
		//fmt.Println("Order registrerd on all peers, Confirmed!")
		return 2
	case isPresent[2] && isPresent[0]: // alle har 0 eller 2 (noen har utført ordren)
		fmt.Println("Order taken")
		return 0

	case isPresent[0] && isPresent[1]: // alle har 0 eller 1 (noen har fått en ny ordre)
		fmt.Println("Order registrerd")
		return 1

	}
	return 0 //default
}

// returnerer true hvis alle heiser holder samme liste med hallrequests
func allSynced(m map[string]nettverk.InformationElev) bool {
	var firstElev nettverk.InformationElev
	isFirst := true

	for _, elev := range m {
		if isFirst {
			firstElev.HallRequests = elev.HallRequests
			isFirst = false
		} else {
			for i := 0; i < len(m); i++ {
				firstPair := firstElev.HallRequests[i]
				for _, s := range elev.HallRequests[1:] {
					if s != firstPair {
						return false
					}
				}
			}
		}
	}
	return true
}
