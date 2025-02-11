package main

type ElevatorInputDevice struct {
	floorSensor   func() int
	requestButton func(int, int) int
	stopButton    func() int
	obstruction   func() int
}

type ElevatorOutputDevice struct {
	floorIndicator     func(int)
	requestButtonLight func(int, int, int)
	doorLight          func(int)
	stopButtonLight    func(int)
	motorDirection     func(int)
}

func wrapRequestButton(f int, b int) int {
	return elevatorHardwareGetButtonSignal(b, f)
}

func wrapRequestButtonLight(f int, b int, v int) {
	elevatorHardwareSetButtonLamp(b, f, v)
}

func wrapMotorDirection(d int) {
	elevatorHardwareSetMotorDirection(d)
}

func elevioGetInputDevice() ElevatorInputDevice {
	return ElevatorInputDevice{
		floorSensor:   elevatorHardwareGetFloorSensorSignal,
		requestButton: wrapRequestButton,
		stopButton:    elevatorHardwareGetStopSignal,
		obstruction:   elevatorHardwareGetObstructionSignal,
	}
}

func elevioGetOutputDevice() ElevatorOutputDevice {
	return ElevatorOutputDevice{
		floorIndicator:     elevatorHardwareSetFloorIndicator,
		requestButtonLight: wrapRequestButtonLight,
		doorLight:          elevatorHardwareSetDoorOpenLamp,
		stopButtonLight:    elevatorHardwareSetStopLamp,
		motorDirection:     wrapMotorDirection,
	}
}

func elevioDirnToString(d int) string {
	switch d {
	case D_Up:
		return "D_Up"
	case D_Down:
		return "D_Down"
	case D_Stop:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func elevioButtonToString(b int) string {
	switch b {
	case B_HallUp:
		return "B_HallUp"
	case B_HallDown:
		return "B_HallDown"
	case B_Cab:
		return "B_Cab"
	default:
		return "B_UNDEFINED"
	}
}
