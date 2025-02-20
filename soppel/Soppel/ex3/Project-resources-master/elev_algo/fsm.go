package main

import (
	"fmt"
)

type Elevator struct {
	floor     int
	dirn      int
	behaviour int
	requests  [N_FLOORS][N_BUTTONS]int
	config    ElevatorConfig
}

type ElevatorConfig struct {
	doorOpenDurationS   float64
	clearRequestVariant int
}

type ElevOutputDevice struct {
	motorDirection     func(int)
	doorLight          func(int)
	floorIndicator     func(int)
	requestButtonLight func(int, int, int)
}

var elevator Elevator
var outputDevice ElevOutputDevice

func fsm_init() {
	elevator = Elevator{}
	elevator.config.doorOpenDurationS = 3.0 // Default value
	outputDevice = elevio_getOutputDevice()
}

func setAllLights(e Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			outputDevice.requestButtonLight(floor, btn, e.requests[floor][btn])
		}
	}
}

func fsm_onInitBetweenFloors() {
	outputDevice.motorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestButtonPress(btnFloor int, btnType int) {
	fmt.Printf("\n\nRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests_shouldClearImmediately(elevator, btnFloor, btnType) {
			timer_start(elevator.config.doorOpenDurationS)
		} else {
			elevator.requests[btnFloor][btnType] = 1
		}
	case EB_Moving:
		elevator.requests[btnFloor][btnType] = 1
	case EB_Idle:
		elevator.requests[btnFloor][btnType] = 1
		pair := requests_chooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case EB_DoorOpen:
			outputDevice.doorLight(1)
			timer_start(elevator.config.doorOpenDurationS)
			elevator = requests_clearAtCurrentFloor(elevator)
		case EB_Moving:
			outputDevice.motorDirection(elevator.dirn)
		}
	}

	setAllLights(elevator)
}

func fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\nFloorArrival(%d)\n", newFloor)

	elevator.floor = newFloor
	outputDevice.floorIndicator(elevator.floor)

	if elevator.behaviour == EB_Moving && requests_shouldStop(elevator) {
		outputDevice.motorDirection(D_Stop)
		outputDevice.doorLight(1)
		elevator = requests_clearAtCurrentFloor(elevator)
		timer_start(elevator.config.doorOpenDurationS)
		setAllLights(elevator)
		elevator.behaviour = EB_DoorOpen
	}
}

func fsm_onDoorTimeout() {
	fmt.Println("\n\nDoorTimeout()")

	if elevator.behaviour == EB_DoorOpen {
		pair := requests_chooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour

		switch elevator.behaviour {
		case EB_DoorOpen:
			timer_start(elevator.config.doorOpenDurationS)
			elevator = requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Moving, EB_Idle:
			outputDevice.doorLight(0)
			outputDevice.motorDirection(elevator.dirn)
		}
	}
}
