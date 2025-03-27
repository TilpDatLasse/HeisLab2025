package fsm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/requests"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
)

var (
	elevator            elev.Elevator
	outputDevice        elev.ElevatorOutputDevice
	motorTimeoutStarted float64 = timer.Get_wall_time()
	ID                  string
)

func FsmInit(id string) {
	elevator = elev.Elevator{}
	elevator.Config.DoorOpenDurationS = 3.0 // Default value
	elevator.Config.ClearRequestVariant = elev.CV_InDirn
	outputDevice = elev.Elevio_getOutputDevice()
	outputDevice.MotorDirection(0)
	elevator.Dirn = 0
	elevator.State = elev.IDLE
	elevator.Obs = false
	elevator.MotorStop = false
	ID = id
	GetCabOrders()
}

func FsmOnInitBetweenFloors() {
	outputDevice.MotorDirection(elev.MD_Down)
	elevator.Dirn = elev.MD_Down
	elevator.State = elev.MOVE
}

func FsmOnRequestButtonPress(btnFloor int, btnType int) {
	if btnType == 2 { //is cab-request
		elevator.Requests[btnFloor][btnType] = 2
		FsmOrderInList(btnFloor, btnType, true)
	} else {
		elevator.Requests[btnFloor][btnType] = 1
		setAllLights(elevator)
	}
}

func FsmOrderInList(btnFloor int, btnType int, isOrder bool) {
	elevator.OwnRequests[btnFloor][btnType] = isOrder // Adding order if there is a new one, deleting if HRA changes its mind
	if !isOrder {                                     //if the HRA says there is no order here, there is nothing else to do (exept deleting it)
		return
	}
	switch elevator.State {
	// case elev.DOOROPEN:
	// 	if requests.ShouldClearImmediately(elevator, btnFloor, btnType) {
	// 		outputDevice.DoorLight(true)
	// 		timer.Timer_start(elevator.Config.DoorOpenDurationS)
	// 		elevator.OwnRequests[btnFloor][btnType] = false
	// 		elevator.Requests[btnFloor][btnType] = 0
	// 		fmt.Println("DEBUG 1")

	// 		FsmOnDoorTimeout()
	// 	} else {
	// 		elevator.OwnRequests[btnFloor][btnType] = true
	// 	}
	case elev.MOVE:
		elevator.OwnRequests[btnFloor][btnType] = true
	case elev.IDLE:
		elevator.OwnRequests[btnFloor][btnType] = true
		elevator.Dirn, elevator.State = requests.ChooseDirection(elevator)

		switch elevator.State {
		case elev.DOOROPEN:
			outputDevice.DoorLight(true)
			timer.Timer_start(elevator.Config.DoorOpenDurationS)
			fmt.Println("DEBUG 2")
			FsmOnDoorTimeout()
			elevator = requests.ClearAtCurrentFloor(elevator)
			SaveCabOrders()
		case elev.MOVE:
			outputDevice.MotorDirection(elev.MotorDirection(elevator.Dirn))
		}
	}

	setAllLights(elevator)
}

func FsmOnFloorArrival(newFloor int) {
	fmt.Println("arrived: ", newFloor)
	motorTimeoutStarted = timer.Get_wall_time()
	elevator.Floor = newFloor
	outputDevice.FloorIndicator(elevator.Floor)

	if elevator.State == elev.MOVE && requests.ShouldStop(elevator) {
		outputDevice.MotorDirection(elev.MD_Stop)
		outputDevice.DoorLight(true)
		elevator = requests.ClearAtCurrentFloor(elevator)
		SaveCabOrders()
		timer.Timer_start(elevator.Config.DoorOpenDurationS)
		setAllLights(elevator)
		elevator.State = elev.DOOROPEN
	}
}

func FsmOnDoorTimeout() {
	if elevator.State == elev.DOOROPEN {
		dirn, behaviour := requests.ChooseDirection(elevator)
		elevator.Dirn = dirn
		elevator.State = behaviour

		switch elevator.State {
		case elev.DOOROPEN:
			timer.Timer_start(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator)
			SaveCabOrders()
			setAllLights(elevator)
		case elev.MOVE, elev.IDLE:
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(elev.MotorDirection(elevator.Dirn))
		}
	}
}

func FsmStop() {
	elev.SetMotorDirection(elev.MD_Stop)
}

func FsmAfterStop() {
	elev.SetMotorDirection(elevator.Dirn)
}

func FlipObs() {
	elevator.Obs = !elevator.Obs
}

func setAllLights(e elev.Elevator) { //trenger egt ikke ta inn elevator her, er global
	for floor := 0; floor < elev.N_FLOORS; floor++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			outputDevice.RequestButtonLight(elev.ButtonType(btn), floor, e.Requests[floor][btn])
		}
	}
}

func FetchElevatorStatus() elev.Elevator {
	return elevator
}

func UpdateHallrequests(hallRequests [][2]elev.ConfirmationState) { // yo her må vi ha cyclicupdate
	for i := 0; i < len(hallRequests); i++ { //for every floor
		for j := 0; j < 2; j++ { //and every button
			list := make([]elev.ConfirmationState, 2)
			list[0] = hallRequests[i][j]
			list[1] = elevator.Requests[i][j]
			//fmt.Println("LISTE: ", list)
			elevator.Requests[i][j] = elev.CyclicUpdate(list, false)
			if elevator.Requests[i][j] == 2 && hallRequests[i][j] != 2 {
				elevator.Requests[i][j] = 1
			}
			if list[1] == 2 && elevator.Requests[i][j] == 0 {
				fmt.Println("--------------------- Order deleted -----------------------")
				fmt.Println("list[0]:", list[0], list[1])
			}
		}
	}
	setAllLights(elevator)
}

func MotorTimeout() {
	var prevState elev.State = elev.IDLE
	timeout_time := 4.0

	for {
		if (elevator.State == elev.MOVE) && (elevator.State != prevState) {
			motorTimeoutStarted = timer.Get_wall_time()

		}
		if (elevator.State == elev.MOVE) && (prevState == elev.MOVE) && ((motorTimeoutStarted + timeout_time) < timer.Get_wall_time()) {
			fmt.Println("---------------------Motor timeout----------------------------")
			elevator.MotorStop = true
			RestartElevator()
			go ifPowerloss()
			motorTimeoutStarted = timer.Get_wall_time()
		}

		prevState = elevator.State

		time.Sleep(200 * time.Millisecond)

	}

}

func ifPowerloss() {

	for elevator.MotorStop {
		if elev.GetFloor() == -1 {
			FsmOnInitBetweenFloors()
		} else {
			FsmInit(ID)
			motorTimeoutStarted = timer.Get_wall_time()
		}
		fmt.Println("motorstop = ", elevator.MotorStop)
		time.Sleep(2 * time.Second)
	}
}

func RestartElevator() { // må vel egt implementere at den sier ifra at den ikke er tilgjengelig
	outputDevice.MotorDirection(elev.MD_Stop)
	for i := 0; i < 300; i++ {
		time.Sleep(10 * time.Millisecond)
		outputDevice.MotorDirection(elev.MD_Stop)
	}
	fmt.Println("Starter heismotor på nytt, går videre")
	elevator.State = elev.IDLE

}

func SaveCabOrders() {
	list := make([]elev.ConfirmationState, 4)

	orderstring := ""
	for i := 0; i < len(elevator.Requests); i++ {
		order := elevator.Requests[i][2] // Assuming the 3rd column holds boolean values
		list[i] = order

	}
	list2 := cabToBool(list)
	for i := 0; i < len(list2); i++ {
		order := list2[i] // Assuming the 3rd column holds boolean values
		str := strconv.FormatBool(order)

		orderstring += (str + " ")
	}

	// Åpner filen for skriving, oppretter den hvis den ikke finnes
	filename := ID + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Feil ved oppretting av fil:", err)
		return
	}
	defer file.Close()

	// Skriver innhold til filen
	_, err = file.WriteString(orderstring)
	if err != nil {
		fmt.Println("Feil ved skriving til fil:", err)
		return
	}

	fmt.Println("Tekst skrevet til fil!")
}

func GetCabOrders() {
	filename := ID + ".txt"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the first line
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()        // Read the line as a string
		words := strings.Fields(line) // Split by whitespace

		// Convert to boolean values
		var bools []bool
		for _, word := range words {
			value, err := strconv.ParseBool(word)
			if err != nil {
				fmt.Println("Error parsing boolean:", err)
				return
			}
			bools = append(bools, value)

		}
		CabRequests := boolToConfirmationState(bools)
		for i := 0; i < len(CabRequests); i++ {
			elevator.Requests[i][2] = CabRequests[i]
			// Assuming the 3rd column holds boolean values

		}
		for i := 0; i < len(bools); i++ {
			elevator.OwnRequests[i][2] = bools[i]
			// Assuming the 3rd column holds boolean values

		}
		// Print the parsed boolean values
		for i := 0; i < len(elevator.OwnRequests); i++ {
			if elevator.OwnRequests[i][2] {
				FsmOrderInList(i, 2, true)
			}

			// Assuming the 3rd column holds boolean values

		}

		fmt.Println("Parsed booleans:", bools)
	}

	// Check for any scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func cabToBool(list []elev.ConfirmationState) []bool {
	boolList := make([]bool, len(list))
	for i, v := range list {
		boolList[i] = v != 0
	}
	return boolList
}

func boolToConfirmationState(list []bool) []elev.ConfirmationState {
	stateList := make([]elev.ConfirmationState, len(list))
	for i, v := range list {
		if v {
			stateList[i] = 2
		} else {
			stateList[i] = 0
		}
	}
	return stateList
}
