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


