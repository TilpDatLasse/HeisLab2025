package syncing

import (
	"fmt"
	"reflect"
	"time"

	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

func Syncing(ch_shouldSync chan bool, ch_fromSync chan map[string]worldview.InformationElev) {
	syncRequest := <-ch_shouldSync
	if syncRequest {
		fmt.Println("Recieved sync request")
		worldview.ShouldSync = true
		go Sync(ch_shouldSync)
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
		fmt.Println("Locked: ", worldview.InfoElev.Locked)

	}
}

func Sync(ch_shouldSync chan bool) {

	for {
		worldview.CompareAndUpdateInfoMap()
		if AllWorldViewsEqual(worldview.WorldViewMap) {
			ch_shouldSync <- false
			fmt.Println("All worldviews are equal")
			break
		} else {
			fmt.Println("Worldviews are not equal")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func AllWorldViewsEqual(worldViewMap map[string]worldview.WorldView) bool {
	worldview.WVMapMutex.Lock() // Lås mutex før lesing fra InfoMap
	defer worldview.WVMapMutex.Unlock()

	var reference *worldview.WorldView
	isFirst := true

	for _, worldView := range worldViewMap {
		if isFirst {
			reference = &worldView
			isFirst = false
			continue
		}

		if !reflect.DeepEqual(reference.InfoMap, worldView.InfoMap) {
			return false
		}
	}

	// OBS: denne kan få koden til å kræsje men nødvendig for å sjekke om alle peers har låst infoen sin for synking
	wv := worldViewMap[worldview.ID] //getting our own infomap
	for id, elev := range wv.InfoMap {
		if elev.Locked != 2 {
			fmt.Printf("Elevator with ID %s is not locked (Locked=%d)\n", id, elev.Locked)
			return false
		}
	}

	return true
}
