package worldview

import (
	"fmt"
	"sync"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
)

var ID string
var InfoMap = make(map[string]InformationElev)
var InfoMapMutex sync.Mutex
var WVMapMutex sync.Mutex

var MyWorldView = WorldView{
	InfoMap: make(map[string]InformationElev), // Initialiserer mappet
}

var WorldViewMap = make(map[string]WorldView) // map som holder alle sine wvs
var ShouldSync bool = false
var InfoElev = InformationElev{
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

func WorldViewFunc(ch_WVRx chan WorldView, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, ch_shouldSync chan bool, id string) {
	ID = id
	MyWorldView.Id = ID
	InfoElev.ElevID = ID

	for {
		wv := <-ch_WVRx //worldview mottatt (dette skjer bare når vi holder på å synke)

		if wv.Id != "" {
			WVMapMutex.Lock()
			WorldViewMap[wv.Id] = wv //oppdaterer wvmap med dens info
			WVMapMutex.Unlock()
			if wv.Id != ID { //noen andre sendte
				InfoMapMutex.Lock()
				InfoMap[wv.Id] = wv.InfoMap[wv.Id] //oppdaterer infoen den sendte om seg selv
				InfoMapMutex.Unlock()
				CompareAndUpdateInfoMap(ch_syncRequestsSingleElev)
			} else {
				// InfoMapMutex.Lock()
				// //InfoMap = wv.InfoMap
				// InfoMapMutex.Unlock()
				// CompareAndUpdateInfoMap()
				// myWorldView.InfoMap = InfoMap
			}
			// InfoMapMutex.Lock()
			// //ch_syncRequestsSingleElev <- InfoMap[ID].HallRequests //må nok bruke mutex her, for å sikre at ikke noen sender noe før vi har fått endret sigle elevs hallrequests
			// time.Sleep(10 * time.Millisecond)
			time.Sleep(10 * time.Millisecond)

			if wv.InfoMap[wv.Id].Locked != 0 && !ShouldSync { //hvis mottar at noen vil synke for første gang
				ch_shouldSync <- true
			}
		}
	}

}

func CompareAndUpdateInfoMap(ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	InfoMapMutex.Lock() // Lås mutex før lesing og skriving til InfoMap
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
				update := elev.CyclicUpdate(listOfElev)

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
		update := elev.CyclicUpdate(listOfElev)

		for key, e := range InfoMap {
			e.Locked = update
			InfoMap[key] = e
			if key == ID {
				InfoElev.Locked = update
			}

		}
	}
	select {
	case ch_syncRequestsSingleElev <- InfoMap[ID].HallRequests:
	default:
	}
	InfoMapMutex.Unlock()
}

// henter status fra heisen og sender på channel som en informationElev-variabel
func SetElevatorStatus(ch_WVTx chan WorldView) {
	for {
		if InfoElev.Locked == 0 { //hvis ikke låst
			InfoElev = Converter(fsm.FetchElevatorStatus())
			if ShouldSync {
				InfoElev.Locked = 1
			}
		}
		if ID != "" {
			InfoElev.ElevID = ID
			InfoMapMutex.Lock()
			InfoMap[ID] = InfoElev
			InfoMapMutex.Unlock()
			MyWorldView.InfoMap = InfoMap
			WVMapMutex.Lock()
			WorldViewMap[ID] = MyWorldView
			WVMapMutex.Unlock()

			select {
			case ch_WVTx <- MyWorldView:
			default:
				fmt.Println("Advarsel: Mistet en WorldViewmelding (kanal full)")
			}

		}
		// InfoMapMutex.Lock()
		// select {
		// case ch_syncRequestsSingleElev <- InfoMap[ID].HallRequests:
		// default:
		// }
		// InfoMapMutex.Unlock()
		time.Sleep(30 * time.Millisecond)

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
