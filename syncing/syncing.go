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

func SyncingMain(ch SyncChans) {
	for {
		SyncRequest = <-ch.ShouldSync
		if SyncRequest { //syncRequest == true, request recieved from HRA or other peer
			worldview.ShouldSync = true
			Sync(ch.ShouldSync, ch.SyncRequestSingleElev)

		} else { //syncRequest == false, sync completed
			fmt.Println("Sync done!!")
			worldview.InfoMapMutex.Lock()
			select {
			case ch.InformationElevFromSync <- worldview.InfoMap:
			default:
				fmt.Println("Warning: message not sent to HRA (channel full)")
			}
			worldview.InfoMapMutex.Unlock()

			worldview.ShouldSync = false //må egt sjekke at de andre har fått sendt før vi låser opp, ellers kan input blir endret før det sendes til hra men virker so det går bra
			worldview.InfoElev.Locked = 0
		}
	}
}

// Compares worldviews until they are all equal
func Sync(ch_shouldSync chan bool, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	for {
		worldview.WVMapMutex.Lock()
		worldview.CompareAndUpdateInfoMap(ch_syncRequestsSingleElev)
		worldview.WVMapMutex.Unlock()
		worldview.WVMapMutex.Lock()
		if AllWorldViewsEqual(worldview.WorldViewMap) {
			go syncDone(ch_shouldSync)
			break
		}
		time.Sleep(300 * time.Millisecond) //prøve å tune denne?
	}
}

func syncDone(ch_shouldSync chan bool) {
	ch_shouldSync <- false
}

// Compares the worldviews of all peers
func AllWorldViewsEqual(worldViewMap map[string]worldview.WorldView) bool {
	worldview.WVMapMutex.Unlock()
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

	// OBS: denne kan få koden til å kræsje men nødvendig for å sjekke om alle peers har låst infoen sin for synking
	wv := worldViewMap[worldview.ID]
	for _, elev := range wv.InfoMap {
		if elev.Locked != 2 {
			return false
		}
	}
	return true
}
