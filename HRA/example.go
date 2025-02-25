package HRA

import (
	"encoding/json"
	"fmt"
	"os/exec"

	//"github.com/TilpDatLasse/HeisLab2025/nettverk"
	//elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	//b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase



func HRAMain() {

	

	

	fmt.Println("kjører HRA")
	hraExecutable := "hall_request_assigner"

	fmt.Println("2")

	jsonBytes, err := json.Marshal(input)
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
