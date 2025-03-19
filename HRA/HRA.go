package HRA

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/nettverk"
)

// type HRAElevState struct {
// 	Behavior    string `json:"behaviour"`
// 	Floor       int    `json:"floor"`
// 	Direction   string `json:"direction"`
// 	CabRequests []bool `json:"cabRequests"`
// }

// type HRAInput struct {
// 	HallRequests [][2]bool               `json:"hallRequests"`
// 	States       map[string]HRAElevState `json:"states"`
// }

func HRAMain(HRAOut chan map[string][][2]bool, ch_shouldSync chan bool, ch_fromSync chan map[string]nettverk.InformationElev) {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	for {
		
		// fmt.Printf("InfoMap: ")
		// for k, v := range nettverk.InfoMap {
		// 	fmt.Printf("%6v :  %+v\n", k, v.HallRequests)
		// }

		time.Sleep(4000 * time.Millisecond)

		
		ch_shouldSync <- true  //forespørsel om synking

		infoMap := <- ch_fromSync  //venter på at synking er ferdig

		

		var input nettverk.HRAInput
		input.States = make(map[string]nettverk.HRAElevState)

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
			HRAOut <- *output //send HRA-output videre
			fmt.Printf("output: \n")
			for k, v := range *output {
				fmt.Printf("%6v :  %+v\n", k, v)
			}
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
