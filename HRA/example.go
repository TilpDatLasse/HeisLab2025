package HRA

import (
	"encoding/json"
	"fmt"
	"os/exec"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

// heihei

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func BroadcastElevatorStatus(ch_HRAInputTx chan HRAInput, statusElev chan elev.Elevator) {
	/*	rawInput := <-statusElev

		input := HRAInput{
			HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
			}
	*/

	b.Transmitter(18000, ch_HRAInputTx)
	fmt.Println("yo")
}

func RecieveElevatorStatus(ch_HRAInputRx chan HRAInput) {
	b.Receiver(18000, ch_HRAInputRx)

}

func HRAMain(ch_HRAInputTx chan HRAInput, ch_HRAInputRx chan HRAInput, change_ch chan bool, ElevatorStatusCh chan elev.Elevator) {
	go RecieveElevatorStatus(ch_HRAInputRx)
	go BroadcastElevatorStatus(ch_HRAInputRx, ElevatorStatusCh)

	/*
		//dette er bare et eksempel på input, skal egt hente fra en kanal
		input := HRAInput{
			HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
			States: map[string]HRAElevState{
				"one": HRAElevState{
					Behavior:    "moving",
					Floor:       2,
					Direction:   "up",
					CabRequests: []bool{false, false, false, true},
				},
				"two": HRAElevState{
					Behavior:    "idle",
					Floor:       0,
					Direction:   "stop",
					CabRequests: []bool{false, false, false, false},
				},
			},
		}

		for {
			select {
			case a := <-change_ch:
				fmt.Println(a, "change")
				HRA(input, ch_HRAInputTx)
			case a := <-ch_HRAInputRx:
				HRA(a, ch_HRAInputTx)
			}
		}*/

	fmt.Println("kjører HRA")
	hraExecutable := "hall_request_assigner"

	fmt.Println("2")

	jsonBytes, err := json.Marshal( /*input*/ )
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
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

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

}
