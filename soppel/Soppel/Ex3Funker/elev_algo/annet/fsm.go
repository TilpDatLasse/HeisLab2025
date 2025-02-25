package annet

import (
	"fmt"
)

var elevator Elevator
var outputDevice ElevatorOutputDevice

func Fsm_init() {
	elevator = Elevator{}
	elevator.config.doorOpenDurationS = 3.0 // Default value
	outputDevice = elevio_getOutputDevice()
}

func setAllLights(e Elevator) {
	fmt.Println("lys")
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
	elevator.behaviour = EB_Moving
}

func Fsm_onRequestButtonPress(btnFloor int, btnType int) {
	fmt.Printf("\n\nRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests_shouldClearImmediately(elevator, btnFloor, btnType) {
			timer_start(elevator.config.doorOpenDurationS)
		} else {
			elevator.requests[btnFloor][btnType] = true
		}
	case EB_Moving:
		elevator.requests[btnFloor][btnType] = true
	case EB_Idle:
		elevator.requests[btnFloor][btnType] = true
		elevator.dirn, elevator.behaviour = requests_chooseDirection(elevator)

		switch elevator.behaviour {
		case EB_DoorOpen:
			outputDevice.doorLight(true)
			timer_start(elevator.config.doorOpenDurationS)
			elevator = requests_clearAtCurrentFloor(elevator)
		case EB_Moving:
			outputDevice.motorDirection(MotorDirection(elevator.dirn))
		}
	}

	setAllLights(elevator)
}

func Fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\nFloorArrival(%d)\n", newFloor)

	elevator.floor = newFloor
	outputDevice.floorIndicator(elevator.floor)

	if elevator.behaviour == EB_Moving && requests_shouldStop(elevator) {
		outputDevice.motorDirection(MD_Stop)
		outputDevice.doorLight(true)
		elevator = requests_clearAtCurrentFloor(elevator)
		timer_start(elevator.config.doorOpenDurationS)
		setAllLights(elevator)
		elevator.behaviour = EB_DoorOpen
	}
}

func Fsm_onDoorTimeout() {
	fmt.Println("\n\nDoorTimeout()")

	if elevator.behaviour == EB_DoorOpen {
		dirn, behaviour := requests_chooseDirection(elevator)
		elevator.dirn = dirn
		elevator.behaviour = behaviour

		switch elevator.behaviour {
		case EB_DoorOpen:
			timer_start(elevator.config.doorOpenDurationS)
			elevator = requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Moving, EB_Idle:
			outputDevice.doorLight(false)
			outputDevice.motorDirection(MotorDirection(elevator.dirn))
		}
	}
}
