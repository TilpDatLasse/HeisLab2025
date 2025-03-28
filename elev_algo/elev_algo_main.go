package elev_algo

import (
	"fmt"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
)

type SingleElevatorChans struct {
	DrvButtons      chan elev.ButtonEvent
	DrvFloors       chan int
	DrvObstr        chan bool
	DrvStop         chan bool
	TimerChannel    chan bool
	SingleElevQueue chan [][2]bool
}

func ElevMain(ch SingleElevatorChans, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, simPort string, id string) {
	elev.Init("localhost:"+simPort, elev.N_FLOORS)
	fsm.FsmInit(id)
	input := elev.Elevio_getInputDevice()

	if input.FloorSensor() == -1 { // If elevator is between floors
		fsm.FsmOnInitBetweenFloors()
		fmt.Println("Pushing Elevator down to closest floor")
	}

	go elev.PollButtons(ch.DrvButtons)
	go elev.PollFloorSensor(ch.DrvFloors)
	go elev.PollObstructionSwitch(ch.DrvObstr)
	go elev.PollStopButton(ch.DrvStop)
	go fsm.MotorTimeout()
	go fsm.ObstructionTimeout()
	go timer.Time(ch.TimerChannel)

	for {
		select {
		case a := <-ch.DrvButtons: // Button press
			fsm.FsmOnRequestButtonPress(a.Floor, int(a.Button))
			fsm.SaveCabOrders()

		case a := <-ch.DrvFloors: // Floor arrival
			fsm.FsmOnFloorArrival(a)

		case <-ch.TimerChannel: // Timer timeout (door close)
			timer.Timer_stop()
			fsm.FsmOnDoorTimeout()

		case <-ch.DrvObstr: // Obstruction switch
			fsm.FlipObs()

		case a := <-ch.DrvStop: // Stop button
			if a {
				fsm.FsmStop()
			}
			if !a {
				fsm.FsmAfterStop()
			}

		case outputHRA := <-ch.SingleElevQueue: // Hall requests from HRA
			for f, floor := range outputHRA {
				for d, isOrder := range floor {
					fsm.FsmOrderInList(f, d, isOrder)
				}
			}

		case hallRequest := <-ch_syncRequestsSingleElev: //Update hall requests
			fsm.UpdateHallrequests(hallRequest)
		}
	}
}
