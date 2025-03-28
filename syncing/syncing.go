package syncing

import (
	"reflect"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

type SyncChans struct {
	ShouldSync              chan bool
	InformationElevFromSync chan map[string]worldview.InformationElev
	SyncRequestSingleElev   chan [][2]elev.ConfirmationState
}

// recieves sync-requests from the HRA and starts the syncing process
func SyncingMain(syncChans SyncChans) {
	for {
		worldview.ShouldSync = <-syncChans.ShouldSync
		Sync(syncChans)
	}
}

// gets updated WorldViewMap from the worldview-module and passes on worldview to the HRA when all are synced
func Sync(syncChans SyncChans) {
	for {
		worldview.CompareAndUpdateInfoMap(syncChans.SyncRequestSingleElev)
		worldview.WVMapMutex.Lock()
		WVMapCopy := worldview.DeepCopyWVMap(worldview.WorldViewMap)
		worldview.WVMapMutex.Unlock()
		if AllWorldViewsEqual(WVMapCopy) {
			syncChans.InformationElevFromSync <- WVMapCopy[worldview.ID].InfoMap
			worldview.ShouldSync = false
			worldview.InfoElev.Locked = 0
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
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
