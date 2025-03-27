package syncing

import (
	"fmt"
	"reflect"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

//var SyncRequest = false

type SyncChans struct {
	ShouldSync              chan bool
	InformationElevFromSync chan map[string]worldview.InformationElev
	SyncRequestSingleElev   chan [][2]elev.ConfirmationState
}

func SyncingMain(syncChans SyncChans, getMyWorldView chan worldview.MyWVrequest, getWorldViewMap chan worldview.WVMapRequest) {

	for {
		SyncRequest := <-syncChans.ShouldSync
		if SyncRequest { //syncRequest == true, request of synching recieved from HRA
			worldview.ShouldSync = true
			//worldview.InfoElev.Locked = 1
			Sync(syncChans, getMyWorldView, getWorldViewMap)

		} //else { //syncRequest == false, sync completed
		// fmt.Println("Sync done!!")

		// select {
		// case syncChans.InformationElevFromSync <- worldview.GetMyWorldView(getMyWorldView).InfoMap:
		// default:
		// 	fmt.Println("Warning: message not sent to HRA (channel full)")
		// }

		//worldview.ShouldSync = false
		// worldview.InfoElev.Locked = 0
		// }
	}
}

func Sync(syncChans SyncChans, getMyWorldView chan worldview.MyWVrequest, getWorldViewMap chan worldview.WVMapRequest) {
	for {
		//worldview.WVMapMutex.Lock()
		worldview.CompareAndUpdateInfoMap(syncChans.SyncRequestSingleElev, getMyWorldView, getWorldViewMap)
		//worldview.WVMapMutex.Unlock()
		//worldview.WVMapMutex.Lock()

		wvMap := worldview.GetWorldViewMap(getWorldViewMap)
		if AllWorldViewsEqual(wvMap) {
			syncChans.InformationElevFromSync <- worldview.GetMyWorldView(getMyWorldView).InfoMap
			//go syncDone(ShouldSync)
			fmt.Println("Sync done!!")
			worldview.InfoElev.Locked = 0
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// func syncDone(ShouldSync chan bool) {
// 	ShouldSync <- false
// }

// Compares the worldviews of all peers
func AllWorldViewsEqual(worldViewMap map[string]worldview.WorldView) bool {
	//worldview.WVMapMutex.Unlock()
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
