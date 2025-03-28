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
	ID         string //Id of local peer
	ShouldSync bool   = false
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
	State              HRAElevState
	HallRequests       [][2]elev.ConfirmationState
	Locked             elev.ConfirmationState
	ElevID             string
	MotorStop          bool
	ObstructionFailure bool
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

// Updates worldview with relevant recieved information
func WorldViewMain(ch_WVRx chan WorldView, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, ch_shouldSync chan bool, id string) {
	ID = id
	MyWorldView.Id = ID
	InfoElev.ElevID = ID
	MyWorldView.PeerList.Id = ID

	for {
		wv := <-ch_WVRx 
		if wv.Id != "" {
			WVMapMutex.Lock()
			WorldViewMap[wv.Id] = wv 
			peers.PeerToUpdate <- wv.PeerList
			WVMapMutex.Unlock()
			if wv.Id != ID { 
				InfoMapMutex.Lock()
				InfoMap[wv.Id] = wv.InfoMap[wv.Id] 
				InfoMapMutex.Unlock()
				CompareAndUpdateInfoMap(ch_syncRequestsSingleElev)
				MyWorldView.Timestamp = timer.Get_wall_time()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}


// Getting the worldview of the local elevator and sending on channel to udp-broadcast
func SetElevatorStatus(ch_WVTx chan WorldView) {
	for {
		peers.PeerToUpdate <- MyWorldView.PeerList
		hasMotorStopped := Converter(fsm.FetchElevatorStatus()).MotorStop
		hasObstructionFailure := Converter(fsm.FetchElevatorStatus()).ObstructionFailure
		if InfoElev.Locked == 0 { 
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
			if hasMotorStopped || hasObstructionFailure {
				continue
			}

			select {
			case ch_WVTx <- deepCopyWV(MyWorldView):
			default:
				fmt.Println("Advarsel: Mistet en WorldViewmelding (kanal full)")
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Comparing info from the different peers to check if we can update the cyclic counters
func CompareAndUpdateInfoMap(ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	wasTimedOut := wasTimedOut()
	infoMap := deepCopyWV(MyWorldView).InfoMap
	if len(infoMap) != 0 {

		//Comparing hallrequests
		for i := 0; i < elev.N_FLOORS; i++ {
			for j := 0; j < elev.N_BUTTONS-1; j++ {
				listOfElev := make([]elev.ConfirmationState, len(infoMap))
				k := 0
				for _, elev := range infoMap {
					if elev.HallRequests == nil {
						return
					}
					listOfElev[k] = elev.HallRequests[i][j]
					k++
				}
				update := elev.CyclicUpdate(listOfElev, wasTimedOut)

				if _, ok := infoMap[ID]; ok {
					InfoMapMutex.Lock()
					InfoMap[ID].HallRequests[i][j] = update
					InfoMapMutex.Unlock()
				}
			}
		}
		// Comparing the Locked-attribute
		listOfElev := make([]elev.ConfirmationState, len(infoMap))
		k := 0
		for _, elev := range infoMap {
			listOfElev[k] = elev.Locked
			k++
		}
		update := elev.CyclicUpdate(listOfElev, wasTimedOut)
		InfoElev.Locked = update
		InfoMapMutex.Lock()
		InfoMap[ID] = InfoElev
		infoMap[ID] = InfoElev
		InfoMapMutex.Unlock()
	}
	select {
	case ch_syncRequestsSingleElev <- infoMap[ID].HallRequests:
	default:
	}
}

// Checks if a peer has been offline, and may have outdated information
func wasTimedOut() bool {
	var timeOut float64 = 1.0
	var keyList []string
	var maxDiff float64 = 0

	WVMapMutex.Lock()
	WVMapCopy := DeepCopyWVMap(WorldViewMap)
	WVMapMutex.Unlock()

	for key, _ := range WVMapCopy {
		keyList = append(keyList, key)
	}
	for i := 0; i < len(keyList)-1; i++ {
		Diff := math.Abs(WVMapCopy[keyList[i]].Timestamp - WVMapCopy[keyList[i+1]].Timestamp)

		if Diff > maxDiff {
			maxDiff = Diff
		}
	}
	return maxDiff > timeOut
}

// Makes a deep copy of a WorldView-type
func deepCopyWV(original WorldView) WorldView {
	InfoMapMutex.Lock()
	copyMap := make(map[string]InformationElev)
	for key, value := range original.InfoMap {
		copyMap[key] = value
	}
	copy := WorldView{
		InfoMap:   copyMap,
		Id:        original.Id,
		Timestamp: original.Timestamp,
		PeerList:  original.PeerList,
	}
	InfoMapMutex.Unlock()
	return copy
}

// makes a deep copy of a WorldViewMap-type
func DeepCopyWVMap(original map[string]WorldView) map[string]WorldView {
	copy := make(map[string]WorldView)
	for key, value := range original {
		copy[key] = deepCopyWV(value)
	}
	return copy
}
