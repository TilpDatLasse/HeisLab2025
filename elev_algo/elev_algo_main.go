package elev_algo

import (
	"fmt"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
	//"github.com/TilpDatLasse/HeisLab2025/nettverk"
)

type SingleElevatorChans struct {
	Drv_buttons       chan elev.ButtonEvent
	Drv_floors        chan int
	Drv_obstr         chan bool
	Drv_stop          chan bool
	Timer_channel     chan bool
	Single_elev_queue chan [][2]bool
}

func ElevMain(ch SingleElevatorChans, ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, simPort string) {
	elev.Init("localhost:"+simPort, elev.N_FLOORS)

	fsm.FsmInit()

	input := elev.Elevio_getInputDevice()

	if input.FloorSensor() == -1 {
		fsm.FsmOnInitBetweenFloors()
		fmt.Println("Dytter heisen ned til nærmeste etasje")
	}

	go elev.PollButtons(ch.Drv_buttons)
	go elev.PollFloorSensor(ch.Drv_floors)
	go elev.PollObstructionSwitch(ch.Drv_obstr)
	go elev.PollStopButton(ch.Drv_stop)
	go fsm.MotorTimeout()
	go timer.Time(ch.Timer_channel)

	for {
		select {
		case a := <-ch.Drv_buttons:
			fsm.FsmOnRequestButtonPress(a.Floor, int(a.Button))

		case a := <-ch.Drv_floors:
			fsm.FsmOnFloorArrival(a)

		case <-ch.Timer_channel:
			timer.Timer_stop()
			fsm.FsmOnDoorTimeout()

		case <-ch.Drv_obstr:
			fsm.FlipObs()

		case a := <-ch.Drv_stop:
			if a {
				fsm.FsmStop()
			}
			if !a {
				fsm.FsmAfterStop()
			}

		case outputHRA := <-ch.Single_elev_queue:
			for f, floor := range outputHRA {
				for d, isOrder := range floor {
					fsm.FsmOrderInList(f, d, isOrder)
				}
			}

		case hallRequest := <-ch_syncRequestsSingleElev:
			fsm.UpdateHallrequests(hallRequest)
		}
	}
}
