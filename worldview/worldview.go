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

type WVUDPChans struct {
	WorldViewTxChan chan WorldView
	WorldViewRxChan chan WorldView
}

type WVServerChans struct {
	GetMyWorldView  chan MyWVrequest
	SetMyWorldView  chan InformationElev
	GetWorldViewMap chan WVMapRequest
	SetWorldViewMap chan WorldView
}

type MyWVrequest struct {
	ResponseChan chan WorldView
}

type WVMapRequest struct {
	ResponseChan chan map[string]WorldView
}

type WorldView struct {
	InfoMap   map[string]InformationElev
	Id        string
	Timestamp float64
	PeerList  peers.PeerUpdate
}

type InformationElev struct {
	State        HRAElevState
	HallRequests [][2]elev.ConfirmationState
	Locked       elev.ConfirmationState
	ElevID       string
	MotorStop    bool
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

func wvMapServer(wvServerChans WVServerChans) {
	for {
		select {
		case wvMapRequest := <-wvServerChans.GetWorldViewMap:
			wvMapRequest.ResponseChan <- WorldViewMap
		case wv := <-wvServerChans.SetWorldViewMap:
			WorldViewMap[wv.Id] = wv
			peers.PeerToUpdate <- wv.PeerList
		}
		time.Sleep(4 * time.Millisecond)
	}
}

func wvServer(wvServerChans WVServerChans) {
	for {
		select {
		case wvRequest := <-wvServerChans.GetMyWorldView:
			wvRequest.ResponseChan <- MyWorldView
		case elevInfo := <-wvServerChans.SetMyWorldView:
			MyWorldView.InfoMap[ID] = elevInfo
		}
		time.Sleep(4 * time.Millisecond)
	}
}

func WorldViewMain(wvServerChans WVServerChans, ch_WVRx chan WorldView, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, ch_shouldSync chan bool, id string) {
	ID = id
	MyWorldView.Id = ID
	InfoElev.ElevID = ID
	MyWorldView.PeerList.Id = ID

	go wvServer(wvServerChans)
	go wvMapServer(wvServerChans)

	for {
		wv := <-ch_WVRx

		if wv.Id != "" {

			wvServerChans.SetWorldViewMap <- wv
			if wv.Id != ID { //noen andre sendte
				responseChan := make(chan WorldView)
				var request MyWVrequest
				request.ResponseChan = responseChan
				wvServerChans.GetMyWorldView <- request
				myWV := <-responseChan
				infoMap := myWV.InfoMap
				infoMap[wv.Id] = wv.InfoMap[wv.Id] //oppdaterer infoen den sendte om seg selv

				CompareAndUpdateInfoMap(ch_syncRequestsSingleElev, wvServerChans.GetMyWorldView, wvServerChans.GetWorldViewMap)
				MyWorldView.Timestamp = timer.Get_wall_time()
				time.Sleep(20 * time.Millisecond)

				// //denne er egt ikke nødvendig, alle synker hele tiden uansett
				// // if wv.InfoMap[wv.Id].Locked != 0 && !ShouldSync { //hvis mottar at noen vil synke for første gang
				// // 	select {
				// // 	case ch_shouldSync <- true:
				// // 	default:
				// // 	}

				// }
			}
		}
	}
}

// Comparing info from the different peers to check if we can update the cyclic counters
func CompareAndUpdateInfoMap(ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, getMyWorldView chan MyWVrequest, getWorldViewMap chan WVMapRequest) {
	infoMap := GetMyWorldView(getMyWorldView).InfoMap

	wasTimedOut := wasTimedOut(getWorldViewMap)
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
					infoMap[ID].HallRequests[i][j] = update
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
		infoMap[ID] = InfoElev

	}
	select {
	case ch_syncRequestsSingleElev <- infoMap[ID].HallRequests:
	default:
	}
}

// henter status fra heisen og sender på channel som en informationElev-variabel
// Getting the status of the local elevator and sending on wv-channel
func SetElevatorStatus(ch_WVTx chan WorldView, wvServerChans WVServerChans) {
	for {
		hasMotorStopped := Converter(fsm.FetchElevatorStatus()).MotorStop
		if InfoElev.Locked == 0 { //hvis ikke låst
			InfoElev = Converter(fsm.FetchElevatorStatus())
			if ShouldSync {
				InfoElev.Locked = 1
			}
		}
		if ID != "" {
			InfoElev.ElevID = ID

			wvServerChans.SetMyWorldView <- InfoElev

			myWV := GetMyWorldView(wvServerChans.GetMyWorldView)

			wvServerChans.SetWorldViewMap <- myWV
			if hasMotorStopped {
				continue
			}
			select {
			case ch_WVTx <- myWV:
			default:
				fmt.Println("Advarsel: Mistet en WorldViewmelding (kanal full)")
			}
		}
		time.Sleep(40 * time.Millisecond)
	}
}

func wasTimedOut(getWorldViewMap chan WVMapRequest) bool {
	var timeOut float64 = 1.0
	var keyList []string
	var maxDiff float64 = 0
	wvMap := GetWorldViewMap(getWorldViewMap)
	for key, _ := range wvMap {
		keyList = append(keyList, key)
	}
	for i := 0; i < len(keyList)-1; i++ {
		Diff := math.Abs(wvMap[keyList[i]].Timestamp - wvMap[keyList[i+1]].Timestamp)
		if Diff > maxDiff {
			maxDiff = Diff
		}
	}
	return maxDiff > timeOut
}

// returns the current local worldView via the vwServer
func GetMyWorldView(GetMyWorldView chan MyWVrequest) WorldView {
	responseChan := make(chan WorldView)
	var request MyWVrequest
	request.ResponseChan = responseChan
	GetMyWorldView <- request
	myWV := <-responseChan
	responseChan = nil
	return myWV
}

// returns the current worldView-map via the vwServer
func GetWorldViewMap(GetWorldViewMap chan WVMapRequest) map[string]WorldView {
	responseChan := make(chan map[string]WorldView)
	var request WVMapRequest
	request.ResponseChan = responseChan
	GetWorldViewMap <- request
	myWV := <-responseChan
	responseChan = nil
	return myWV
}
