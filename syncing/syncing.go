package syncing

import (
	"reflect"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	wv "github.com/TilpDatLasse/HeisLab2025/worldview"
)

type SyncChans struct {
	ShouldSync              chan bool
	InformationElevFromSync chan map[string]wv.InformationElev
	SyncRequestSingleElev   chan [][2]elev.ConfirmationState
}

// recieves sync-requests from the HRA and starts the syncing process
func SyncingMain(syncChans SyncChans) {
	for {
		wv.ShouldSync = <-syncChans.ShouldSync
		Sync(syncChans)
	}
}

// gets updated WorldViewMap from the worldview-module and passes on worldview to the HRA when all are synced
func Sync(syncChans SyncChans) {
	for {
		wv.CompareAndUpdateInfoMap(syncChans.SyncRequestSingleElev)
		wv.WVMapMutex.Lock()
		WVMapCopy := wv.DeepCopyWVMap(wv.WorldViewMap)
		wv.WVMapMutex.Unlock()
		if AllWorldViewsEqual(WVMapCopy) {
			syncChans.InformationElevFromSync <- WVMapCopy[wv.ID].InfoMap
			wv.ShouldSync = false
			wv.InfoElev.Locked = 0
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// Compares the worldviews of all peers
func AllWorldViewsEqual(worldViewMap map[string]wv.WorldView) bool {
	var reference wv.WorldView
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
	wv := worldViewMap[wv.ID]
	for _, elev := range wv.InfoMap {
		if elev.Locked != 2 {
			return false
		}
	}
	return true
}
