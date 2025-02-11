package main

import (
	"fmt"
	"time"
)

type ElevInputDevice struct {
	floorSensor   func() int
	requestButton func(int, int) int
}

func elevio_getInputDevice() ElevInputDevice {
	return ElevInputDevice{
		floorSensor:   func() int { return -1 },
		requestButton: func(f, b int) int { return 0 },
	}
}

/*
func fsm_onInitBetweenFloors() {
    fmt.Println("FSM initialized between floors")
}

func fsm_onRequestButtonPress(floor, button int) {
    fmt.Printf("Request button pressed at floor %d, button %d\n", floor, button)
}

func fsm_onFloorArrival(floor int) {
    fmt.Printf("Elevator arrived at floor %d\n", floor)
}

func fsm_onDoorTimeout() {
    fmt.Println("Door timeout occurred")
}

func timer_timedOut() bool {
    return false
}

func timer_stop() {
    fmt.Println("Timer stopped")
}

*/

func main() {
	fmt.Println("Started!")

	inputPollRateMs := 25
	input := elevio_getInputDevice()

	if input.floorSensor() == -1 {
		fsm_onInitBetweenFloors()
	}

	prevRequest := make([][]int, N_FLOORS)
	for i := range prevRequest {
		prevRequest[i] = make([]int, N_BUTTONS)
	}

	prevFloor := -1

	for {
		// Request button
		for f := 0; f < N_FLOORS; f++ {
			for b := 0; b < N_BUTTONS; b++ {
				v := input.requestButton(f, b)
				if v != 0 && v != prevRequest[f][b] {
					fsm_onRequestButtonPress(f, b)
				}
				prevRequest[f][b] = v
			}
		}

		// Floor sensor
		f := input.floorSensor()
		if f != -1 && f != prevFloor {
			fsm_onFloorArrival(f)
		}
		prevFloor = f

		// Timer
		if timer_timedOut() {
			timer_stop()
			fsm_onDoorTimeout()
		}

		time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
	}
}
