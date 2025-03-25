package syncing

import (
	"fmt"
	"reflect"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

func Syncing(ch_shouldSync chan bool, ch_fromSync chan map[string]worldview.InformationElev, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	for {
		syncRequest := <-ch_shouldSync
		if syncRequest {
			worldview.ShouldSync = true
			Sync(ch_shouldSync, ch_syncRequestsSingleElev) //channel will be blocking if this is not a go-routine

		} else { //syncRequest == false, synk ferdig
			fmt.Println("Sync done!!")
			worldview.InfoMapMutex.Lock() // Lås mutex før lesing fra InfoMap
			select {
			case ch_fromSync <- worldview.InfoMap:
			default:
				fmt.Println("Advarsel: Mistet en infomapmelding (kanal full)")
			}
			worldview.InfoMapMutex.Unlock() // Lås opp mutex etter lesing

			worldview.ShouldSync = false //må egt sjekke at de andre har fått sendt før vi låser opp
			worldview.InfoElev.Locked = 0
			//fmt.Println("Locked: ", worldview.InfoElev.Locked)

		}
	}

}

func Sync(ch_shouldSync chan bool, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState) {
	for {
		worldview.WVMapMutex.Lock()
		worldview.CompareAndUpdateInfoMap(ch_syncRequestsSingleElev, false) //wasTimedOut can only be true when we receive a new update
		worldview.WVMapMutex.Unlock()
		if AllWorldViewsEqual(worldview.WorldViewMap) {
			go syncDone(ch_shouldSync)
			//fmt.Println("All worldviews are equal")
			break
		} else {
			fmt.Println("Worldviews are not equal")
			//fmt.Println("WV: ", worldview.WorldViewMap)
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func syncDone(ch_shouldSync chan bool) {
	ch_shouldSync <- false
}

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

	// OBS: denne kan få koden til å kræsje men nødvendig for å sjekke om alle peers har låst infoen sin for synking
	wv := worldViewMap[worldview.ID] //getting our own infomap
	for _, elev := range wv.InfoMap {
		if elev.Locked != 2 {
			//fmt.Printf("Elevator with ID %s is not locked (Locked=%d)\n", id, elev.Locked)
			//fmt.Println("InfoMap: ", worldview.InfoMap)
			return false
		}
	}

	return true
}
