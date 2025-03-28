package fsm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/requests"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
)

var (
	elevator     elev.Elevator
	outputDevice elev.ElevatorOutputDevice
	ID           string
)

func FsmInit(id string) {
	elevator = elev.Elevator{}
	elevator.Config.DoorOpenDurationS = 3.0
	elevator.Config.ClearRequestVariant = elev.CV_InDirn
	outputDevice = elev.Elevio_getOutputDevice()
	outputDevice.MotorDirection(0)
	elevator.Dirn = 0
	elevator.State = elev.IDLE
	elevator.Obs = false
	elevator.MotorStop = false
	elevator.ObstructionFailure = false
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
		setAllLights()
	}
}

func FsmOrderInList(btnFloor int, btnType int, isOrder bool) {
	elevator.OwnRequests[btnFloor][btnType] = isOrder // Adding order if there is a new one, deleting if HRA changes its mind
	if !isOrder {                                     //if the HRA says there is no order here, there is nothing else to do (exept deleting it)
		return
	}
	switch elevator.State {

	case elev.MOVE:
		elevator.OwnRequests[btnFloor][btnType] = true
	case elev.IDLE:
		elevator.OwnRequests[btnFloor][btnType] = true
		elevator.Dirn, elevator.State = requests.ChooseDirection(elevator)

		switch elevator.State {
		case elev.DOOROPEN:
			outputDevice.DoorLight(true)
			timer.Timer_start(elevator.Config.DoorOpenDurationS)
			FsmOnDoorTimeout()
			elevator = requests.ClearAtCurrentFloor(elevator)
			SaveCabOrders()
		case elev.MOVE:
			outputDevice.MotorDirection(elev.MotorDirection(elevator.Dirn))
		}
	}

	setAllLights()
}

func FsmOnFloorArrival(newFloor int) {
	motorTimeoutStarted = timer.Get_wall_time()
	elevator.Floor = newFloor
	outputDevice.FloorIndicator(elevator.Floor)

	if elevator.State == elev.MOVE && requests.ShouldStop(elevator) {
		outputDevice.MotorDirection(elev.MD_Stop)
		outputDevice.DoorLight(true)
		elevator = requests.ClearAtCurrentFloor(elevator)
		SaveCabOrders()
		timer.Timer_start(elevator.Config.DoorOpenDurationS)
		setAllLights()
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
			setAllLights()
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

func setAllLights() {
	for floor := 0; floor < elev.N_FLOORS; floor++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			outputDevice.RequestButtonLight(elev.ButtonType(btn), floor, elevator.Requests[floor][btn])
		}
	}
}

func FetchElevatorStatus() elev.Elevator {
	return elevator
}

func UpdateHallrequests(hallRequests [][2]elev.ConfirmationState) {
	for i := 0; i < len(hallRequests); i++ { //for every floor
		for j := 0; j < 2; j++ { //and every button
			list := make([]elev.ConfirmationState, 2)
			list[0] = hallRequests[i][j]
			list[1] = elevator.Requests[i][j]
			elevator.Requests[i][j] = elev.CyclicUpdate(list, false)
			if elevator.Requests[i][j] == 2 && hallRequests[i][j] != 2 {
				elevator.Requests[i][j] = 1
			}
			if list[1] == 2 && elevator.Requests[i][j] == 0 {
				fmt.Println("--------------------- Order deleted -----------------------") //Is printed when a different peer clears an order
			}
		}
	}
	setAllLights()
}

// Saves the cab orders to a txt-file
func SaveCabOrders() {
	list := make([]elev.ConfirmationState, 4)

	orderstring := ""
	for i := 0; i < len(elevator.Requests); i++ {
		order := elevator.Requests[i][2]
		list[i] = order

	}
	list2 := cabToBool(list)
	for i := 0; i < len(list2); i++ {
		order := list2[i]
		str := strconv.FormatBool(order)

		orderstring += (str + " ")
	}
	filename := ID + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("An error occured while opening the file", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(orderstring)
	if err != nil {
		fmt.Println("Error while writing to file: ", err)
		return
	}
}

// Gets the cab orders from a txt-file
func GetCabOrders() {
	filename := ID + ".txt"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)

		var bools []bool
		for _, word := range words {
			value, err := strconv.ParseBool(word)
			if err != nil {
				fmt.Println("Error parsing boolean:", err)
				return
			}
			bools = append(bools, value)

		}
		fmt.Println("Parsed booleans:", bools)
		CabRequests := boolToConfirmationState(bools)
		for i := 0; i < len(CabRequests); i++ {
			elevator.Requests[i][2] = CabRequests[i]

		}
		for i := 0; i < len(bools); i++ {
			elevator.OwnRequests[i][2] = bools[i]

		}
		for i := 0; i < len(elevator.OwnRequests); i++ {
			if elevator.OwnRequests[i][2] {

				FsmOrderInList(i, 2, true)

			}

		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

// Converts a list of ConfirmationStates to a list of bools
func cabToBool(list []elev.ConfirmationState) []bool {
	boolList := make([]bool, len(list))
	for i, v := range list {
		boolList[i] = v != 0
	}
	return boolList
}

// Converts a list of bools to a list of ConfirmationStates
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
