package nettverk

import (
	"fmt"
	"os"
	"reflect"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/localip"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
)

var ID string
var InfoMap = make(map[string]InformationElev)

var myWorldView = WorldView{
	InfoMap: make(map[string]InformationElev), // Initialiserer mappet
}

var WorldViewMap = make(map[string]WorldView) // map som holder alle sine wvs
// trenger ikke bruke denne som heartbeat, kan informationElev til det
var shouldSync bool = false
var infoElev InformationElev

type WorldView struct {
	InfoMap map[string]InformationElev
	Id      string
}

type InformationElev struct {
	State        HRAElevState
	HallRequests [][2]elev.ConfirmationState // denne skal deles med alle peers, så alle vet hvilke ordre som er aktive
	Locked       elev.ConfirmationState      // Når denne er !=0 skal ikke lenger info hentes fra elev-modulen
	ID           string
}

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

// henter status fra heisen og sender på channel som en informationElev-variabel
func SetElevatorStatus(ch_HRAInputTx chan InformationElev, ch_WVTx chan WorldView) {
	for {
		//fmt.Println("setelevatorstatus ", infoElev.ID)
		//info := Converter(fsm.FetchElevatorStatus()) // skal egt ikke være her
		if infoElev.Locked == 0 {
			infoElev = Converter(fsm.FetchElevatorStatus())
			if shouldSync {
				infoElev.Locked = 1
			}
		}
		infoElev.ID = ID
		myWorldView.Id = ID
		select {
		case ch_HRAInputTx <- infoElev:
			//fmt.Println("DEBUG 1", infoElev)
		default:
			fmt.Println("Advarsel: Mistet en elevatorstatusmelding (kanal full)")
		}
		select {
		case ch_WVTx <- myWorldView:
			//fmt.Println("DEBUG 1", infoElev)
		default:
			fmt.Println("Advarsel: Mistet en WorldViewmelding (kanal full)")
		}
		//fmt.Println("DEBUG 2")
		time.Sleep(1000 * time.Millisecond)
	}
}

func BroadcastElevatorStatus(ch_HRAInputTx chan InformationElev) {
	b.Transmitter(14000, ch_HRAInputTx)
}

func RecieveElevatorStatus(ch_HRAInputRx chan InformationElev) {
	b.Receiver(14000, ch_HRAInputRx)
}

func BroadcastWV(ch_WVTx chan WorldView) {
	b.Transmitter(15000, ch_WVTx)
}

func RecieveWV(ch_WVRx chan WorldView) {
	b.Receiver(15000, ch_WVRx)
}

func Nettverk_hoved(ch_HRAInputRx chan InformationElev, ch_WVRx chan WorldView, ch_shouldSync chan bool, ch_fromSync chan map[string]InformationElev, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, id string) {

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	ID = id
	//myWorldView.id = ID

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(16000, id, peerTxEnable)
	go peers.Receiver(16000, peerUpdateCh)

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-ch_HRAInputRx: //heartbeat med info mottatt
			fmt.Println("Recieved heartbeat: ", a.HallRequests)
			//fmt.Println("InfoMap: ")
			//for k, v := range InfoMap {
			//fmt.Printf("%6v :  %+v\n", k, v.HallRequests)
			//}
			if a.ID != "" {
				InfoMap[a.ID] = a
				myWorldView.InfoMap = InfoMap
				CompareAndUpdateInfoMap()
				ch_syncRequestsSingleElev <- InfoMap[ID].HallRequests //må nok bruke mutex her, for å sikre at ikke noen sender noe før vi har fått endret sigle elevs hallrequests
				if a.Locked != 0 && !shouldSync {                     //hvis mottar at noen vil synke for første gang
					shouldSync = true
					//SetElevatorStatus(ch_HRAInputRx) //henter info siste gang før synk og låser info
					go Sync(ch_shouldSync)
				}
			}
		case syncRequest := <-ch_shouldSync:
			if syncRequest {
				shouldSync = true
				//SetElevatorStatus(ch_HRAInputRx) //henter info siste gang før synk og låser info
				go Sync(ch_shouldSync)
			} else { //syncRequest == false, synk ferdig
				ch_fromSync <- InfoMap
				shouldSync = false //må egt sjekke at de andre har fått sendt før vi låser opp
			}

		case wv := <-ch_WVRx: //worldview mottatt (dette skjer bare når vi holder på å synke)
			fmt.Println("Recieved WorldView from id: ", wv.Id)
			if wv.Id != "" {
				WorldViewMap[wv.Id] = wv
			}
		}

	}
}

// konverterer en elev.elevator-variabel til en InformationElev-variabel
func Converter(e elev.Elevator) InformationElev {
	rawInput := e
	hallRequests := make([][2]elev.ConfirmationState, len(rawInput.Requests))
	cabRequests := make([]elev.ConfirmationState, len(rawInput.Requests))

	for i := 0; i < len(rawInput.Requests); i++ {
		hallRequests[i] = [2]elev.ConfirmationState{rawInput.Requests[i][0], rawInput.Requests[i][1]}
		cabRequests[i] = rawInput.Requests[i][2]
	}

	input := InformationElev{
		HallRequests: hallRequests,
		State: HRAElevState{
			Behavior:    stateToString(rawInput.State),
			Floor:       rawInput.Floor,
			Direction:   dirnToString(rawInput.Dirn),
			CabRequests: cabToBool(cabRequests),
		},
	}
	return input
}

// konverterer states vi bruker til states HRA bruker
func stateToString(s elev.State) string {
	switch s {
	case elev.IDLE:
		return "idle"
	case elev.MOVE:
		return "moving"
	case elev.DOOROPEN:
		return "doorOpen"
	case elev.STOP:
		return "doorOpen"
	default:
		return "idle"
	}
}

// konverterer dirn vi bruker til dirn HRA bruker
func dirnToString(s elev.MotorDirection) string {
	switch s {
	case elev.MD_Up:
		return "up"
	case elev.MD_Down:
		return "down"
	case elev.MD_Stop:
		return "stop"
	default:
		return "stop"
	}
}

func cabToBool(list []elev.ConfirmationState) []bool {
	boolList := make([]bool, len(list))
	for i, v := range list {
		boolList[i] = v != 0 // Convert non-zero values to true, zero to false
	}

	return boolList
}

// Henter output fra HRA og sender videre til elev-modulen
func FromHRA(HRAOut chan map[string][][2]bool, ch_elevator_queue chan [][2]bool) {
	for {
		output := <-HRAOut
		for k, v := range output {
			if k == ID {
				ch_elevator_queue <- v
			}
		}
	}
}

// fra sync

func Sync(ch_shouldSync chan bool) {
	for {
		if AllWorldViewsEqual(WorldViewMap) {
			ch_shouldSync <- false
			fmt.Println("All worldviews are equal")
			break
		} else {
			fmt.Println("Worldviews are not equal")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func CompareAndUpdateInfoMap() {
	//infoMap := <-ch_toSync
	if len(InfoMap) != 0 {
		//denne delen sammenlikner hallrequests og oppdaterer de
		for i := 0; i < elev.N_FLOORS; i++ {
			for j := 0; j < elev.N_BUTTONS-1; j++ {
				listOfElev := make([]elev.ConfirmationState, len(InfoMap))
				k := 0
				for _, elev := range InfoMap {
					listOfElev[k] = elev.HallRequests[i][j]
					k++
				}
				update := cyclicUpdate(listOfElev)

				for _, elev := range InfoMap {
					elev.HallRequests[i][j] = update
				}
			}
		}
		// denne delen sammenlikner locked og oppdaterer
		listOfElev := make([]elev.ConfirmationState, len(InfoMap))
		k := 0
		for _, elev := range InfoMap {
			listOfElev[k] = elev.Locked
			k++
		}
		update := cyclicUpdate(listOfElev)

		for key, elev := range InfoMap {
			elev.Locked = update
			InfoMap[key] = elev
		}
	}
	//ch_fromSync <- infoMap
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
		//fmt.Println("Order registrerd on all peers, Confirmed!")
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

func AllWorldViewsEqual(worldViewMap map[string]WorldView) bool {
	var reference *WorldView
	isFirst := true

	for _, worldView := range worldViewMap {
		if isFirst {
			reference = &worldView
			isFirst = false
			continue
		}

		if !reflect.DeepEqual(reference.InfoMap, worldView.InfoMap) {
			return false
		}
	}

	
	// OBS: denne får koden til å kræsje men nødvendig for å sjekke om alle peers har låst infoen sin for synking
	// wv := worldViewMap[ID]
	// for id, elev := range wv.InfoMap {
	// 	if elev.Locked != 2 {
	// 		fmt.Printf("Elevator with ID %s is not locked (Locked=%d)\n", id, elev.Locked)
	// 		return false
	// 	}
	// }

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

func PrintTest() {
	for {
		fmt.Println("InfoMap: ")
		for k, v := range InfoMap {
			fmt.Printf("%6v :  %+v\n", k, v.HallRequests)
		}
		time.Sleep(5000 * time.Millisecond)
	}
}
