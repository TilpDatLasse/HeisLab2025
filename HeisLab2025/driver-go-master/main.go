package main

import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	state_channel := make(chan elevio.State)

	elevio.SetButtonLampsOff()

	go elevio.SetState(state_channel)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			elevio.SetFloorIndicator(a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}
			state_channel <- elevio.STOP // denne linjen blokkerer

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case a := <-state_channel:
			fmt.Printf("state changed %v", a)
			if a == elevio.INIT {

			}
			if a == elevio.IDLE {

			}
			if a == elevio.MOVE {

			}
			if a == elevio.STOP {

			}
			if a == elevio.DOOROPEN {

			}
		}
	}
}

// tips til oss selv
// Vi får ikke statesa til å fungere enda
// Kanskje set_state() blokkerer, siden kanalen ikke er en buffer og kun kan inneholde en verdi av gangen.
// deilen ligger i at ingen henter informasjonen fra states kanalen. Ingen hente og da blokkerer den.
