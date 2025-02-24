package elev_algo

import (
	"fmt"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
	//"github.com/TilpDatLasse/HeisLab2025/nettverk"
)

func Elevator_hoved(buttonTx chan elev.ButtonEvent, buttonRx chan elev.ButtonEvent, change_ch chan bool) {
	fmt.Println("Started!")

	elev.Init("localhost:15657", elev.N_FLOORS)

	fsm.Fsm_init()

	input := elev.Elevio_getInputDevice()

	if input.FloorSensor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
	}

	drv_buttons := make(chan elev.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	timer_channel := make(chan bool)

	go elev.PollButtons(drv_buttons)
	go elev.PollFloorSensor(drv_floors)
	go elev.PollObstructionSwitch(drv_obstr)
	go elev.PollStopButton(drv_stop)
	go timer.Time(timer_channel)

	for {
		select {
		case a := <-drv_buttons:
			
			if a.Button != elev.BT_Cab {
				buttonTx <- a
			} else {
				fsm.Fsm_onRequestButtonPress(a.Floor, int(a.Button))
			}
			change_ch <- true

		case a := <-buttonRx:
			fmt.Printf("Received for call: %#v\n", a)
			fsm.Fsm_onRequestButtonPress(a.Floor, int(a.Button))

		case a := <-drv_floors:
			fsm.Fsm_onFloorArrival(a)

		case a := <-timer_channel:
			timer.Timer_stop()
			fmt.Println(a)
			fsm.Fsm_onDoorTimeout()

		case a := <-drv_obstr:
			fmt.Println(a)
			fsm.FlipObs()

		case a := <-drv_stop:

			if a {
				fsm.Fsm_stop()
			}
			if !a {
				fsm.Fsm_after_stop()

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
