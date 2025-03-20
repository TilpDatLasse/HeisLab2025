package nettverk

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/localip"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
)

var ID string
var InfoMap = make(map[string]InformationElev)
var infoMapMutex sync.Mutex
var WVMapMutex sync.Mutex

var myWorldView = WorldView{
	InfoMap: make(map[string]InformationElev), // Initialiserer mappet
}

var WorldViewMap = make(map[string]WorldView) // map som holder alle sine wvs
var shouldSync bool = false
var infoElev = InformationElev{
	ElevID: ID,
}

type WorldView struct {
	InfoMap map[string]InformationElev
	Id      string
}

type InformationElev struct {
	State        HRAElevState
	HallRequests [][2]elev.ConfirmationState // denne skal deles med alle peers, så alle vet hvilke ordre som er aktive
	Locked       elev.ConfirmationState      // Når denne er !=0 skal ikke lenger info hentes fra elev-modulen
	ElevID       string
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
		if infoElev.Locked == 0 { //hvis ikke låst
			infoElev = Converter(fsm.FetchElevatorStatus())
			if shouldSync {
				infoElev.Locked = 1
				fmt.Println("debug 1")
			}
		}
		if ID != "" {
			infoElev.ElevID = ID
			InfoMap[ID] = infoElev
			myWorldView.InfoMap = InfoMap

			select {
			case ch_WVTx <- myWorldView:
			default:
				fmt.Println("Advarsel: Mistet en WorldViewmelding (kanal full)")
			}
			time.Sleep(1000 * time.Millisecond)
		}

	}
}

func BroadcastWV(ch_WVTx chan WorldView) {
	b.Transmitter(14500, ch_WVTx)
}

func RecieveWV(ch_WVRx chan WorldView) {
	b.Receiver(14500, ch_WVRx)
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
	infoElev.ElevID = ID
	myWorldView.Id = ID

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

		case syncRequest := <-ch_shouldSync:

			if syncRequest {
				fmt.Println("Recieved sync request")
				shouldSync = true
				go Sync(ch_shouldSync)
			} else { //syncRequest == false, synk ferdig
				fmt.Println("Sync done!!")
				infoMapMutex.Lock() // Lås mutex før lesing fra InfoMap
				select {
				case ch_fromSync <- InfoMap:
				default:
					fmt.Println("Advarsel: Mistet en infomapmelding (kanal full)")
				}
				infoMapMutex.Unlock() // Lås opp mutex etter lesing

				shouldSync = false //må egt sjekke at de andre har fått sendt før vi låser opp
				infoElev.Locked = 0
				fmt.Println("Locked: ", infoElev.Locked)

			}

		case wv := <-ch_WVRx: //worldview mottatt (dette skjer bare når vi holder på å synke)

			if wv.Id != "" {
				WVMapMutex.Lock()
				WorldViewMap[wv.Id] = wv //oppdaterer wvmap med dens info
				WVMapMutex.Unlock()
				if wv.Id != ID {
					infoMapMutex.Lock()
					InfoMap[wv.Id] = wv.InfoMap[wv.Id] //oppdaterer infoen den sendte om seg selv
					infoMapMutex.Unlock()
					CompareAndUpdateInfoMap()
				} else { //det var vi som sendte, vi kan nå oppdatere vår egen info
					infoMapMutex.Lock()
					//InfoMap = wv.InfoMap
					infoMapMutex.Unlock()
					CompareAndUpdateInfoMap()
					myWorldView.InfoMap = InfoMap
				}
				infoMapMutex.Lock()
				ch_syncRequestsSingleElev <- InfoMap[ID].HallRequests //må nok bruke mutex her, for å sikre at ikke noen sender noe før vi har fått endret sigle elevs hallrequests
				time.Sleep(100 * time.Millisecond)
				infoMapMutex.Unlock()

				if wv.InfoMap[wv.Id].Locked != 0 && !shouldSync { //hvis mottar at noen vil synke for første gang
					shouldSync = true
					go Sync(ch_shouldSync)
				}
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
		time.Sleep(1000 * time.Millisecond)
	}
}

// fra sync

func Sync(ch_shouldSync chan bool) {
	for {
		CompareAndUpdateInfoMap()
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
	infoMapMutex.Lock() // Lås mutex før lesing og skriving til InfoMap
	defer infoMapMutex.Unlock()
	//fmt.Println("INFO MAP: ", InfoMap)
	if len(InfoMap) != 0 {
		//denne delen sammenlikner hallrequests og oppdaterer de
		for i := 0; i < elev.N_FLOORS; i++ {
			for j := 0; j < elev.N_BUTTONS-1; j++ {
				listOfElev := make([]elev.ConfirmationState, len(InfoMap))
				//fmt.Println("LEN: ", len(InfoMap))
				k := 0
				for _, elev := range InfoMap {
					if elev.HallRequests == nil {
						return
					}
					//fmt.Println("ELEV: ", elev.HallRequests)
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

		for key, e := range InfoMap {
			e.Locked = update
			InfoMap[key] = e
			if key == ID {
				infoElev.Locked = update
			}

		}
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
		//fmt.Println("Order registrerd on all peers, Confirmed!")
		return 2
	case isPresent[2] && isPresent[0]: // alle har 0 eller 2 (noen har utført ordren)
		return 0
	case isPresent[0] && isPresent[1]: // alle har 0 eller 1 (noen har fått en ny ordre)
		return 1
	}
	return 0 //default
}

func AllWorldViewsEqual(worldViewMap map[string]WorldView) bool {
	WVMapMutex.Lock() // Lås mutex før lesing fra InfoMap
	defer WVMapMutex.Unlock()

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

	// OBS: denne kan få koden til å kræsje men nødvendig for å sjekke om alle peers har låst infoen sin for synking
	wv := worldViewMap[ID]
	for id, elev := range wv.InfoMap {
		if elev.Locked != 2 {
			fmt.Printf("Elevator with ID %s is not locked (Locked=%d)\n", id, elev.Locked)
			return false
		}
	}

	return true
}
