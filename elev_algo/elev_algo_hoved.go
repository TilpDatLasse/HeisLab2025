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

func Elevator_hoved(ch SingleElevatorChans) {
	fmt.Println("Started!")

	elev.Init("localhost:15657", elev.N_FLOORS)

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

			//fmt.Printf("Received for call: %#v\n", a)

		case a := <-ch.Drv_floors:
			fsm.Fsm_onFloorArrival(a)

		case a := <-ch.Timer_channel:
			timer.Timer_stop()
			fmt.Println(a)
			fsm.Fsm_onDoorTimeout()

		case a := <-ch.Drv_obstr:
			fmt.Println(a)
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
				for d, _ := range floor {
					fsm.Fsm_OrderInList(f, d)
				}

			}
		}
	}

	/*
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


			time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
		}
	*/
}
