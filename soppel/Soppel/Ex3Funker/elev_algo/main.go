package main

import (
	"elev_algo/annet"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Started!")

	annet.Init("localhost:15657", annet.N_FLOORS)

	annet.Fsm_init()
	annet.SetButtonLamp(1, 1, true)
	time.Sleep(10000)

	inputPollRateMs := 25
	input := annet.Elevio_getInputDevice()

	if input.FloorSensor() == -1 {
		annet.Fsm_onInitBetweenFloors()
	}

	prevRequest := make([][]bool, annet.N_FLOORS)
	for i := range prevRequest {
		prevRequest[i] = make([]bool, annet.N_BUTTONS)
	}

	prevFloor := -1

	for {
		for f := 0; f < annet.N_FLOORS; f++ {
			for b := 0; b < annet.N_BUTTONS; b++ {
				v := input.RequestButton(annet.ButtonType(b), f)
				if v {
					annet.Fsm_onRequestButtonPress(f, b)
				}
			}
		}

		// Floor sensor
		f := input.FloorSensor()
		if f != -1 && f != prevFloor {
			annet.Fsm_onFloorArrival(f)
		}
		prevFloor = f

		// Timer
		if annet.Timer_timed_out() {
			annet.Timer_stop()
			annet.Fsm_onDoorTimeout()
		}

		time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
	}
}
