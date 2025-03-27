package fsm

import (
	"fmt"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/requests"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
)

var (
	elevator            elev.Elevator
	outputDevice        elev.ElevatorOutputDevice
	motorTimeoutStarted float64 = timer.Get_wall_time()
)

func FsmInit() {
	elevator.Config.DoorOpenDurationS = 3.0
	elevator.Config.ClearRequestVariant = elev.CV_InDirn
	outputDevice = elev.Elevio_getOutputDevice()
	outputDevice.MotorDirection(0)
	elevator.Dirn = 0
	elevator.State = elev.IDLE
	elevator.Obs = false
	elevator.MotorStop = false
}

func FsmOnInitBetweenFloors() {
	outputDevice.MotorDirection(-1)
	elevator.Dirn = -1
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
	if !isOrder {                                     //If the HRA says there is no order here, there is nothing else to do (exept deleting it)
		return
	}
	switch elevator.State {
	case elev.DOOROPEN:
		if requests.ShouldClearImmediately(elevator, btnFloor, btnType) {
			elevator.OwnRequests[btnFloor][btnType] = false
			elevator.Requests[btnFloor][btnType] = 0
			timer.Timer_start(elevator.Config.DoorOpenDurationS)
			FsmOnDoorTimeout()
		} else {
			elevator.OwnRequests[btnFloor][btnType] = true
		}
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
		for j := 0; j < 2; j++ { //for every button
			list := make([]elev.ConfirmationState, 2)
			list[0] = hallRequests[i][j]
			list[1] = elevator.Requests[i][j]
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
	setAllLights()
}

func MotorTimeout() {
	var prevState elev.State = elev.IDLE
	timeoutTime := 4.0
	for {
		if (elevator.State == elev.MOVE) && (elevator.State != prevState) {
			motorTimeoutStarted = timer.Get_wall_time()

		}
		// Checks if the elevator has been moving for too long without reaching its destination
		if (elevator.State == elev.MOVE) && (prevState == elev.MOVE) && ((motorTimeoutStarted + timeoutTime) < timer.Get_wall_time()) {
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

func RestartElevator() { // må vel egt implementere at den sier ifra at den ikke er tilgjengelig
	outputDevice.MotorDirection(elev.MD_Stop)
	for i := 0; i < 800; i++ {
		time.Sleep(10 * time.Millisecond)
		outputDevice.MotorDirection(elev.MD_Stop)
	}
	elevator.State = elev.IDLE

}

func ifPowerloss() {

	for elevator.MotorStop {
		if elev.GetFloor() == -1 {
			FsmOnInitBetweenFloors()
		} else {
			FsmInit()
			motorTimeoutStarted = timer.Get_wall_time()
		}
		fmt.Println("motorstop = ", elevator.MotorStop)
		time.Sleep(2 * time.Second)
	}
}
