package HRA

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/syncing"
	"github.com/TilpDatLasse/HeisLab2025/worldview"
)

func HRAMain(ch_elevatorQueue chan [][2]bool, ch_shouldSync chan bool, ch_fromSync chan map[string]worldview.InformationElev, ID string) {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	case "darwin":
		hraExecutable = "hall_request_assigner_mac"
	default:
		panic("OS not supported")
	}

	for {

		time.Sleep(1000 * time.Millisecond)

		if !syncing.SyncRequest { // litt dårlig kode men funker vel
			ch_shouldSync <- true //forespørsel om synking
		}

		infoMap := <-ch_fromSync //venter på at synking er ferdig

		var input worldview.HRAInput
		input.States = make(map[string]worldview.HRAElevState)

		for key := range infoMap {
			elevstate := infoMap[key].State
			input.States[key] = elevstate
			input.HallRequests = hallToBool(infoMap[key].HallRequests) //koverterer fra confirmationstate til bool
		}

		if len(infoMap) > 0 {
			jsonBytes, err := json.Marshal(input)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				return
			}

			ret, err := exec.Command("./HRA/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
			if err != nil {
				fmt.Println("exec.Command error: ", err)
				fmt.Println(string(ret))
				return
			}

			output := new(map[string][][2]bool)
			err = json.Unmarshal(ret, &output)
			if err != nil {
				fmt.Println("json.Unmarshal error: ", err)
				return
			}

			sendToElev(*output, ch_elevatorQueue, ID)

			fmt.Printf("output: \n")
			for k, v := range *output {
				fmt.Printf("%6v :  %+v\n", k, v)
			}
		}
	}
}

// Sender output til elev-modulen
func sendToElev(output map[string][][2]bool, ch_elevatorQueue chan [][2]bool, ID string) {
	for k, v := range output {
		if k == ID {
			ch_elevatorQueue <- v
		}
	}

}

func hallToBool(hallReqList [][2]elev.ConfirmationState) [][2]bool {
	boolList := make([][2]bool, len(hallReqList))
	for i, v := range hallReqList {
		boolList[i][0] = v[0] == 2
		boolList[i][1] = v[1] == 2
	}
	return boolList
}
