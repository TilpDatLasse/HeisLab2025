package fsm

import (
	"fmt"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
)

var elevator elev.Elevator
var outputDevice elev.ElevatorOutputDevice

func Fsm_init() {
	elevator = Elevator{}
	elevator.config.doorOpenDurationS = 3.0 // Default value
	elevator.config.clearRequestVariant = CV_InDirn
	outputDevice = elevio_getOutputDevice()
	outputDevice.motorDirection(0)
	elevator.dirn = 0
	elevator.state = IDLE
	elevator.obs = false
}

func setAllLights(e Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			outputDevice.requestButtonLight(ButtonType(btn), floor, e.requests[floor][btn])
		}
	}
}

func Fsm_onInitBetweenFloors() {
	print(outputDevice.motorDirection)
	outputDevice.motorDirection(-1)
	elevator.dirn = -1
	elevator.state = MOVE
}

func Fsm_onRequestButtonPress(btnFloor int, btnType int) {
	fmt.Printf("\n\nRequestButtonPress(%d, %d)\n", btnFloor, btnType)
	fmt.Printf("state(%d)", elevator.state)

	switch elevator.state {
	case DOOROPEN:
		if requests_shouldClearImmediately(elevator, btnFloor, btnType) {
			timer_start(elevator.config.doorOpenDurationS)

			Fsm_onDoorTimeout()
		} else {
			elevator.requests[btnFloor][btnType] = true
		}
	case MOVE:
		elevator.requests[btnFloor][btnType] = true
	case IDLE:
		elevator.requests[btnFloor][btnType] = true
		elevator.dirn, elevator.state = requests_chooseDirection(elevator)

		switch elevator.state {
		case DOOROPEN:
			outputDevice.doorLight(true)
			timer_start(elevator.config.doorOpenDurationS)

			Fsm_onDoorTimeout()
			elevator = requests_clearAtCurrentFloor(elevator)
		case MOVE:
			outputDevice.motorDirection(MotorDirection(elevator.dirn))
		}
	}

	setAllLights(elevator)
}

func Fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\nFloorArrival(%d)\n", newFloor)

	elevator.floor = newFloor
	outputDevice.floorIndicator(elevator.floor)

	if elevator.state == MOVE && requests_shouldStop(elevator) {
		outputDevice.motorDirection(MD_Stop)
		outputDevice.doorLight(true)
		elevator = requests_clearAtCurrentFloor(elevator)
		timer_start(elevator.config.doorOpenDurationS)
		setAllLights(elevator)
		elevator.state = DOOROPEN
	}
}

func Fsm_onDoorTimeout() {
	fmt.Println("\n\nDoorTimeout()")

	if elevator.state == DOOROPEN {
		dirn, behaviour := requests_chooseDirection(elevator)
		elevator.dirn = dirn
		elevator.state = behaviour

		switch elevator.state {
		case DOOROPEN:
			timer_start(elevator.config.doorOpenDurationS)
			elevator = requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case MOVE, IDLE:
			outputDevice.doorLight(false)
			outputDevice.motorDirection(MotorDirection(elevator.dirn))
		}
	}
}

func Fsm_stop() {

	SetMotorDirection(MD_Stop)

}

func Fsm_after_stop() {
	SetMotorDirection(elevator.dirn)

}

func FlipObs() {
	elevator.obs = !elevator.obs
}

//fra requests

func requests_above(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func requests_chooseDirection(e Elevator) (MotorDirection, State) {
	fmt.Println("chooseDir")
	switch e.dirn {
	case MD_Up:
		if requests_above(e) {
			return MD_Up, MOVE
		} else if requests_here(e) {
			return MD_Down, DOOROPEN
		} else if requests_below(e) {
			return MD_Down, MOVE
		}
	case MD_Down:
		if requests_below(e) {
			return MD_Down, MOVE
		} else if requests_here(e) {
			return MD_Up, DOOROPEN
		} else if requests_above(e) {
			return MD_Up, MOVE
		}
	case MD_Stop:
		if requests_here(e) {
			return MD_Stop, DOOROPEN
		} else if requests_above(e) {
			return MD_Up, MOVE
		} else if requests_below(e) {
			return MD_Down, MOVE
		}
	}
	return MD_Stop, IDLE
}

func requests_shouldStop(e Elevator) bool {
	switch e.dirn {
	case MD_Down:
		return e.requests[e.floor][B_HallDown] || e.requests[e.floor][B_Cab] || !requests_below(e)
	case MD_Up:
		return e.requests[e.floor][B_HallUp] || e.requests[e.floor][B_Cab] || !requests_above(e)
	default:
		return true
	}

}

func requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type int) bool {
	switch e.config.clearRequestVariant {
	case CV_All:
		return e.floor == btn_floor
	case CV_InDirn:
		return e.floor == btn_floor &&
			(e.dirn == MD_Up && btn_type == B_HallUp ||
				e.dirn == MD_Down && btn_type == B_HallDown ||
				e.dirn == MD_Stop ||
				btn_type == B_Cab)
	default:
		return false
	}
}

func requests_clearAtCurrentFloor(e Elevator) Elevator {
	switch e.config.clearRequestVariant {
	case CV_All:
		fmt.Println("All")
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.requests[e.floor][btn] = false
		}
	case CV_InDirn:
		fmt.Println("InDirn")
		e.requests[e.floor][B_Cab] = false
		switch e.dirn {
		case MD_Up:
			if !requests_above(e) && !e.requests[e.floor][B_HallUp] {
				e.requests[e.floor][B_HallDown] = false
			}
			e.requests[e.floor][B_HallUp] = false
		case MD_Down:
			if !requests_below(e) && !e.requests[e.floor][B_HallDown] {
				e.requests[e.floor][B_HallUp] = false
			}
			e.requests[e.floor][B_HallDown] = false
		case MD_Stop:
		default:
			e.requests[e.floor][B_HallUp] = false
			e.requests[e.floor][B_HallDown] = false
		}
	}
	return e
}
