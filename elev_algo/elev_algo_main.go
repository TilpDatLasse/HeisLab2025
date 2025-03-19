package elev_algo

import (
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

func Elev_main(ch SingleElevatorChans,ch_syncRequestsSingleElev chan [][2]elev.ConfirmationState, simPort string) {
	elev.Init("localhost:"+simPort, elev.N_FLOORS) //burde bruke flag her for å teste med flere

	fsm.Fsm_init()

	input := elev.Elevio_getInputDevice()

	if input.FloorSensor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
	}

	go elev.PollButtons(ch.Drv_buttons)
	go elev.PollFloorSensor(ch.Drv_floors)
	go elev.PollObstructionSwitch(ch.Drv_obstr)
	go elev.PollStopButton(ch.Drv_stop)
	go timer.Time(ch.Timer_channel)

	for {
		select {
		case a := <-ch.Drv_buttons:
			fsm.Fsm_onRequestButtonPress(a.Floor, int(a.Button))

		case a := <-ch.Drv_floors:
			fsm.Fsm_onFloorArrival(a)

		case <-ch.Timer_channel:
			timer.Timer_stop()
			fsm.Fsm_onDoorTimeout()

		case <-ch.Drv_obstr:
			fsm.FlipObs()

		case a := <-ch.Drv_stop:
			if a {
				fsm.Fsm_stop()
			}
			if !a {
				fsm.Fsm_after_stop()
			}

		case outputHRA := <-ch.Single_elev_queue:
			for f, floor := range outputHRA {
				for d, isOrder := range floor {
					if isOrder {
						fsm.Fsm_OrderInList(f, d)
					}
				}
			}
		case hallRequest := <-ch_syncRequestsSingleElev:
			fsm.UpdateHallrequests(hallRequest)
		
		}

	}
}
