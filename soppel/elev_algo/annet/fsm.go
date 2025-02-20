package annet

import (
	"fmt"
)

var elevator Elevator
var outputDevice ElevatorOutputDevice

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
