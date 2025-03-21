package fsm

import (
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
)

var elevator elev.Elevator
var outputDevice elev.ElevatorOutputDevice

func FetchElevatorStatus() elev.Elevator {
	return elevator
}

func Fsm_init() {
	elevator = elev.Elevator{}
	elevator.Config.DoorOpenDurationS = 3.0 // Default value
	elevator.Config.ClearRequestVariant = elev.CV_InDirn
	outputDevice = elev.Elevio_getOutputDevice()
	outputDevice.MotorDirection(0)
	elevator.Dirn = 0
	elevator.State = elev.IDLE
	elevator.Obs = false
}

func setAllLights(e elev.Elevator) { //trenger egt ikke ta inn elevator her, er global
	for floor := 0; floor < elev.N_FLOORS; floor++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			outputDevice.RequestButtonLight(elev.ButtonType(btn), floor, e.Requests[floor][btn])
		}
	}
}

func Fsm_onInitBetweenFloors() {
	outputDevice.MotorDirection(-1)
	elevator.Dirn = -1
	elevator.State = elev.MOVE
}

func Fsm_onRequestButtonPress(btnFloor int, btnType int) {
	if btnType == 2 { //er cab-request
		elevator.Requests[btnFloor][btnType] = 2
		Fsm_OrderInList(btnFloor, btnType)
	} else {
		elevator.Requests[btnFloor][btnType] = 1
		setAllLights(elevator)
	}
}

func Fsm_OrderInList(btnFloor int, btnType int) {
	elevator.OwnRequests[btnFloor][btnType] = true

	switch elevator.State {
	case elev.DOOROPEN:
		if requests_shouldClearImmediately(elevator, btnFloor, btnType) {
			elevator.OwnRequests[btnFloor][btnType] = false
			elevator.Requests[btnFloor][btnType] = 0
			timer.Timer_start(elevator.Config.DoorOpenDurationS)

			Fsm_onDoorTimeout()
		} else {
			elevator.OwnRequests[btnFloor][btnType] = true
		}
	case elev.MOVE:
		elevator.OwnRequests[btnFloor][btnType] = true
	case elev.IDLE:
		elevator.OwnRequests[btnFloor][btnType] = true
		elevator.Dirn, elevator.State = requests_chooseDirection(elevator)

		switch elevator.State {
		case elev.DOOROPEN:
			outputDevice.DoorLight(true)
			timer.Timer_start(elevator.Config.DoorOpenDurationS)

			Fsm_onDoorTimeout()
			elevator = requests_clearAtCurrentFloor(elevator)
		case elev.MOVE:
			outputDevice.MotorDirection(elev.MotorDirection(elevator.Dirn))
		}
	}

	setAllLights(elevator)
}

func Fsm_onFloorArrival(newFloor int) {
	//fmt.Printf("\n\nFloorArrival(%d)\n", newFloor)

	elevator.Floor = newFloor
	outputDevice.FloorIndicator(elevator.Floor)

	if elevator.State == elev.MOVE && requests_shouldStop(elevator) {
		outputDevice.MotorDirection(elev.MD_Stop)
		outputDevice.DoorLight(true)
		elevator = requests_clearAtCurrentFloor(elevator)
		timer.Timer_start(elevator.Config.DoorOpenDurationS)
		setAllLights(elevator)
		elevator.State = elev.DOOROPEN
	}
}

func Fsm_onDoorTimeout() {
	if elevator.State == elev.DOOROPEN {
		dirn, behaviour := requests_chooseDirection(elevator)
		elevator.Dirn = dirn
		elevator.State = behaviour

		switch elevator.State {
		case elev.DOOROPEN:
			timer.Timer_start(elevator.Config.DoorOpenDurationS)
			elevator = requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case elev.MOVE, elev.IDLE:
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(elev.MotorDirection(elevator.Dirn))
		}
	}
}

func Fsm_stop() {
	elev.SetMotorDirection(elev.MD_Stop)
}

func Fsm_after_stop() {
	elev.SetMotorDirection(elevator.Dirn)
}

func GetObs() bool {
	return elevator.Obs
}

func FlipObs() {
	elevator.Obs = !elevator.Obs
}

//fra requests

func requests_above(e elev.Elevator) bool {
	for f := e.Floor + 1; f < elev.N_FLOORS; f++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			if e.OwnRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e elev.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			if e.OwnRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e elev.Elevator) bool {
	for btn := 0; btn < elev.N_BUTTONS; btn++ {
		if e.OwnRequests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func requests_chooseDirection(e elev.Elevator) (elev.MotorDirection, elev.State) {
	switch e.Dirn {
	case elev.MD_Up:
		if requests_above(e) {
			return elev.MD_Up, elev.MOVE
		} else if requests_here(e) {
			return elev.MD_Down, elev.DOOROPEN
		} else if requests_below(e) {
			return elev.MD_Down, elev.MOVE
		}
	case elev.MD_Down:
		if requests_below(e) {
			return elev.MD_Down, elev.MOVE
		} else if requests_here(e) {
			return elev.MD_Up, elev.DOOROPEN
		} else if requests_above(e) {
			return elev.MD_Up, elev.MOVE
		}
	case elev.MD_Stop:
		if requests_here(e) {
			return elev.MD_Stop, elev.DOOROPEN
		} else if requests_above(e) {
			return elev.MD_Up, elev.MOVE
		} else if requests_below(e) {
			return elev.MD_Down, elev.MOVE
		}
	}
	return elev.MD_Stop, elev.IDLE
}

func requests_shouldStop(e elev.Elevator) bool {
	switch e.Dirn {
	case elev.MD_Down:
		return e.OwnRequests[e.Floor][elev.B_HallDown] || e.OwnRequests[e.Floor][elev.B_Cab] || !requests_below(e)
	case elev.MD_Up:
		return e.OwnRequests[e.Floor][elev.B_HallUp] || e.OwnRequests[e.Floor][elev.B_Cab] || !requests_above(e)
	default:
		return true
	}

}

func requests_shouldClearImmediately(e elev.Elevator, btn_floor int, btn_type int) bool {
	switch e.Config.ClearRequestVariant {
	case elev.CV_All:
		return e.Floor == btn_floor
	case elev.CV_InDirn:
		return e.Floor == btn_floor &&
			(e.Dirn == elev.MD_Up && btn_type == elev.B_HallUp ||
				e.Dirn == elev.MD_Down && btn_type == elev.B_HallDown ||
				e.Dirn == elev.MD_Stop ||
				btn_type == elev.B_Cab)
	default:
		return false
	}
}

func requests_clearAtCurrentFloor(e elev.Elevator) elev.Elevator {
	switch e.Config.ClearRequestVariant {
	case elev.CV_All:
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			e.OwnRequests[e.Floor][btn] = false
		}
	case elev.CV_InDirn:
		e.OwnRequests[e.Floor][elev.B_Cab] = false
		e.Requests[e.Floor][elev.B_Cab] = 0
		switch e.Dirn {
		case elev.MD_Up:
			if !requests_above(e) && !e.OwnRequests[e.Floor][elev.B_HallUp] {
				e.OwnRequests[e.Floor][elev.B_HallDown] = false
				e.Requests[e.Floor][elev.B_HallDown] = 0
			}
			e.OwnRequests[e.Floor][elev.B_HallUp] = false
			e.Requests[e.Floor][elev.B_HallUp] = 0

		case elev.MD_Down:
			if !requests_below(e) && !e.OwnRequests[e.Floor][elev.B_HallDown] {
				e.OwnRequests[e.Floor][elev.B_HallUp] = false
				e.Requests[e.Floor][elev.B_HallUp] = 0
			}
			e.OwnRequests[e.Floor][elev.B_HallDown] = false
			e.Requests[e.Floor][elev.B_HallDown] = 0
		case elev.MD_Stop:
			e.OwnRequests[e.Floor][elev.B_HallUp] = false
			e.Requests[e.Floor][elev.B_HallUp] = 0
			e.OwnRequests[e.Floor][elev.B_HallDown] = false
			e.Requests[e.Floor][elev.B_HallDown] = 0

		default:
			e.OwnRequests[e.Floor][elev.B_HallUp] = false
			e.Requests[e.Floor][elev.B_HallUp] = 0
			e.OwnRequests[e.Floor][elev.B_HallDown] = false
			e.Requests[e.Floor][elev.B_HallDown] = 0
		}
	}
	return e
}

func UpdateHallrequests(hallRequests [][2]elev.ConfirmationState) { // yo her må vi ha cyclicupdate
	for i := 0; i < len(hallRequests); i++ { //itererer over etasjer
		for j := 0; j < 2; j++ { //itererer over buttons/retninger
			list := make([]elev.ConfirmationState, 2)
			list[0] = hallRequests[i][j]
			list[1] = elevator.Requests[i][j]
			//fmt.Println("LISTE: ", list)
			elevator.Requests[i][j] = cyclicUpdate(list)
		}
	}
	setAllLights(elevator)
}

// burde egt definere denne i egen modul så den kan brukes av flere
func cyclicUpdate(list []elev.ConfirmationState) elev.ConfirmationState {
	isPresent := map[elev.ConfirmationState]bool{} // map som lagrer om hver confimationstate(0,1,2) er tilstede
	for _, v := range list {
		isPresent[v] = true
	}
	switch {
	case isPresent[0] && isPresent[1] && isPresent[2]:
		panic("Confirmationstates 0,1,2 at the same time :(")
	case !isPresent[0]: // alle har 1 eller 2
		//fmt.Println("Order registrerd on all peers, Confirmed!")
		return 2
	case isPresent[2] && isPresent[0]: // alle har 0 eller 2 (noen har utført ordren)
		return 0
	case isPresent[0] && isPresent[1]: // alle har 0 eller 1 (noen har fått en ny ordre)
		return 1
	}
	return 0 //default
}
