package HRA

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TilpDatLasse/HeisLab2025/nettverk"
	//elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	//b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	//"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

func getLastSixDigits(str string) int {
	// Del opp strengen på "-" og hent det siste segmentet
	parts := strings.Split(str, "-")
	lastPart := parts[len(parts)-1]
	// Konverter det siste segmentet til et heltall
	num, err := strconv.Atoi(lastPart)
	if err != nil {
		fmt.Println("Feil ved konvertering:", err)
	}
	return num

}

func HRAMain() {

	fmt.Println("kjører HRA")
	//hraExecutable := "hall_request_assigner"

	fmt.Println("2")

	for {

		// Lager en slice for å lagre nøklene
		var keys []string

		// Iterere gjennom map'et og hente nøklene
		for key := range nettverk.InfoMap {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return getLastSixDigits(keys[i]) < getLastSixDigits(keys[j])
		})

		myList := []string{"one", "two", "three"}

		var input nettverk.HRAInput

		for i := 0; i < len(keys); i++ {
			elevstate := nettverk.InfoMap[keys[i]]
			input.States[myList[i]] = elevstate.State
		}
		if len(keys) > 0 {
			input.HallRequests = nettverk.InfoMap[keys[0]].HallRequests

		}

		//fmt.Println(input.States["one"].Floor)
		time.Sleep(1000 * time.Millisecond)
	}
	/*
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
	*/

}
