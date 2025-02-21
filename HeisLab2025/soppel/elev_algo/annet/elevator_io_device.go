package annet

type ElevatorInputDevice struct {
	FloorSensor   func() int
	RequestButton func(ButtonType, int) bool
	stopButton    func() bool
	obstruction   func() bool
}

type ElevatorOutputDevice struct {
	floorIndicator     func(int)
	requestButtonLight func(ButtonType, int, bool)
	doorLight          func(bool)
	stopButtonLight    func(bool)
	motorDirection     func(MotorDirection)
}

func Elevio_getInputDevice() ElevatorInputDevice {
	return ElevatorInputDevice{
		FloorSensor:   GetFloor,
		RequestButton: GetButton,
		stopButton:    GetStop,
		obstruction:   GetObstruction,
	}
}

func elevio_getOutputDevice() ElevatorOutputDevice {
	return ElevatorOutputDevice{
		floorIndicator:     SetFloorIndicator,
		requestButtonLight: SetButtonLamp,
		doorLight:          SetDoorOpenLamp,
		stopButtonLight:    SetStopLamp,
		motorDirection:     SetMotorDirection,
	}
}
