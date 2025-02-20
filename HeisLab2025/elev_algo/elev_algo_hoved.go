package elev_algo

import (
	"fmt"

	"github.com/TilpDatLasse/HeisLab2025/elev_algo/annet"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network_go"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network_go/network/bcast"
)

func Elevator_hoved() {
	fmt.Println("Started!")

	annet.Init("localhost:15657", annet.N_FLOORS)

	annet.Fsm_init()

	input := annet.Elevio_getInputDevice()

	if input.FloorSensor() == -1 {
		annet.Fsm_onInitBetweenFloors()
	}

	drv_buttons := make(chan annet.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	timer_channel := make(chan bool)
	ch := make(chan network_go.HelloMsg)

	go annet.PollButtons(drv_buttons)
	go annet.PollFloorSensor(drv_floors)
	go annet.PollObstructionSwitch(drv_obstr)
	go annet.PollStopButton(drv_stop)
	go annet.Time(timer_channel)
	go bcast.Transmitter(17000, ch)

	for {
		select {
		case a := <-drv_buttons:
			annet.Fsm_onRequestButtonPress(a.Floor, int(a.Button))
			helloMsg := network_go.HelloMsg{"Transmitting" + "2", a.Floor}
			ch <- helloMsg
		case a := <-drv_floors:
			annet.Fsm_onFloorArrival(a)

		case a := <-timer_channel:
			annet.Timer_stop()
			fmt.Println(a)
			annet.Fsm_onDoorTimeout()

		case a := <-drv_obstr:
			fmt.Println(a)
			annet.FlipObs()

		case a := <-drv_stop:

			if a {
				annet.Fsm_stop()
			}
			if !a {
				annet.Fsm_after_stop()

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
