package worldview

import (
	"fmt"
	"math"
	"sync"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
	"github.com/TilpDatLasse/HeisLab2025/network/peers"
)

var (
	InfoMapMutex sync.Mutex
	WVMapMutex   sync.Mutex
)

var (
	ID         string
	ShouldSync bool = false
	InfoElev   InformationElev
)

var (
	InfoMap      = map[string]InformationElev{}
	WorldViewMap = map[string]WorldView{}
	MyWorldView  = WorldView{
		InfoMap:   map[string]InformationElev{},
		Timestamp: timer.Get_wall_time(),
	}
)

type WVChans struct {
	WorldViewTxChan chan WorldView
	WorldViewRxChan chan WorldView
}

type WorldView struct {
	InfoMap   map[string]InformationElev
	Id        string
	Timestamp float64
	PeerList  peers.PeerUpdate
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

func WorldViewMain(ch_WVRx chan WorldView, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, ch_shouldSync chan bool, id string) {
	ID = id
	MyWorldView.Id = ID
	InfoElev.ElevID = ID
	MyWorldView.PeerList.Id = ID

	for {
		wv := <-ch_WVRx //worldview mottatt

		if wv.Id != "" {
			WVMapMutex.Lock()
			WorldViewMap[wv.Id] = wv //oppdaterer wvmap med dens info
			peers.PeerToUpdate <- wv.PeerList
			WVMapMutex.Unlock()
			if wv.Id != ID { //noen andre sendte
				InfoMapMutex.Lock()
				InfoMap[wv.Id] = wv.InfoMap[wv.Id] //oppdaterer infoen den sendte om seg selv
				InfoMapMutex.Unlock()
				CompareAndUpdateInfoMap(ch_syncRequestsSingleElev)
				MyWorldView.Timestamp = timer.Get_wall_time()
				time.Sleep(10 * time.Millisecond)

				//denne er egt ikke nødvendig, alle synker hele tiden uansett
				if wv.InfoMap[wv.Id].Locked != 0 && !ShouldSync { //hvis mottar at noen vil synke for første gang
					select {
					case ch_shouldSync <- true:
					default:
					}

				}
			}
		}
	}
}

// Comparing info from the different peers to check if we can update the cyclic counters
func CompareAndUpdateInfoMap(ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	InfoMapMutex.Lock()
	wasTimedOut := wasTimedOut()
	if len(InfoMap) != 0 {

		//Comparing hallrequests
		for i := 0; i < elev.N_FLOORS; i++ {
			for j := 0; j < elev.N_BUTTONS-1; j++ {
				listOfElev := make([]elev.ConfirmationState, len(InfoMap))
				k := 0
				for _, elev := range InfoMap {
					if elev.HallRequests == nil {
						return
					}
					listOfElev[k] = elev.HallRequests[i][j]
					k++
				}
				update := elev.CyclicUpdate(listOfElev, wasTimedOut)

				if _, ok := InfoMap[ID]; ok {
					InfoMap[ID].HallRequests[i][j] = update
				}
			}
		}

		// Comparing the Locked-attribute
		listOfElev := make([]elev.ConfirmationState, len(InfoMap))
		k := 0
		for _, elev := range InfoMap {
			listOfElev[k] = elev.Locked
			k++
		}
		update := elev.CyclicUpdate(listOfElev, wasTimedOut)

		InfoElev.Locked = update
		InfoMap[ID] = InfoElev

	}
	select {
	case ch_syncRequestsSingleElev <- InfoMap[ID].HallRequests:
	default:
	}
	InfoMapMutex.Unlock()
}

// henter status fra heisen og sender på channel som en informationElev-variabel
// Getting the status of the local elevator and sending on wv-channel
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
		time.Sleep(50 * time.Millisecond)
	}
}

func wasTimedOut() bool {
	var timeOut float64 = 1.0
	var keyList []string
	var maxDiff float64 = 0
	//MÅ HA EN MUTEX HER, men koden kræsjer
	//WVMapMutex.Lock()
	for key, _ := range WorldViewMap {
		keyList = append(keyList, key)
	}
	//WVMapMutex.Unlock()
	for i := 0; i < len(keyList)-1; i++ {
		//WVMapMutex.Lock()
		Diff := math.Abs(WorldViewMap[keyList[i]].Timestamp - WorldViewMap[keyList[i+1]].Timestamp)
		//WVMapMutex.Unlock()
		if Diff > maxDiff {
			maxDiff = Diff
		}
	}
	return maxDiff > timeOut
}
