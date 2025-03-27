package syncing

import (
	"fmt"
	"reflect"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

var SyncRequest = false

type SyncChans struct {
	ShouldSync              chan bool
	InformationElevFromSync chan map[string]worldview.InformationElev
	SyncRequestSingleElev   chan [][2]elev.ConfirmationState
}

func SyncingMain(syncChans SyncChans) {

	for {
		SyncRequest = <-syncChans.ShouldSync
		if SyncRequest { //syncRequest == true, request of synching recieved from HRA or other peer
			worldview.ShouldSync = true
			Sync(syncChans.ShouldSync, syncChans.SyncRequestSingleElev)
			fmt.Println("sted 3")

		} else { //syncRequest == false, sync completed
			fmt.Println("Sync done!!")
			worldview.InfoMapMutex.Lock()
			select {
			case syncChans.InformationElevFromSync <- worldview.InfoMap:
				fmt.Println("sted 2")
			default:
				fmt.Println("Warning: message not sent to HRA (channel full)")
			}
			worldview.InfoMapMutex.Unlock()
			fmt.Println("sted 1")
			worldview.ShouldSync = false
			worldview.InfoElev.Locked = 0
		}
	}
}

func Sync(ShouldSync chan bool, SyncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	for {
		worldview.CompareAndUpdateInfoMap(SyncRequestsSingleElev)
		worldview.WVMapMutex.Lock()
		WVMapCopy := worldview.DeepCopyWVMap(worldview.WorldViewMap)
		worldview.WVMapMutex.Unlock()
		if AllWorldViewsEqual(WVMapCopy) {
			go syncDone(ShouldSync)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func syncDone(ShouldSync chan bool) {
	ShouldSync <- false
}

// Compares the worldviews of all peers
func AllWorldViewsEqual(worldViewMap map[string]worldview.WorldView) bool {
	var reference worldview.WorldView
	isFirst := true

	for _, worldView := range worldViewMap {
		if isFirst {
			reference = worldView
			isFirst = false
			continue
		}
		if !reflect.DeepEqual(reference.InfoMap, worldView.InfoMap) {
			return false
		}
	}

	// Checks if all peers have locked their worldview information before synching
	wv := worldViewMap[worldview.ID]
	for _, elev := range wv.InfoMap {
		if elev.Locked != 2 {
			return false
		}
	}
	return true
}
