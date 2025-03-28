package fsm

import (
	"fmt"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/timer"
)

var (
	motorTimeoutStarted       float64
	ObstructionTimeoutStarted float64
)

func MotorTimeout() {
	var prevState elev.State = elev.IDLE
	timeout_time := 4.0

	for {
		if (elevator.State == elev.MOVE) && (elevator.State != prevState) {
			motorTimeoutStarted = timer.Get_wall_time()

		}
		if (elevator.State == elev.MOVE) && (prevState == elev.MOVE) && ((motorTimeoutStarted + timeout_time) < timer.Get_wall_time()) {
			fmt.Println("---------------------Motor timeout----------------------------")
			elevator.MotorStop = true
			RestartElevator()
			go ifPowerloss()
			motorTimeoutStarted = timer.Get_wall_time()
		}

		prevState = elevator.State

		time.Sleep(200 * time.Millisecond)

	}

}

func ifPowerloss() {

	for elevator.MotorStop {
		if elev.GetFloor() == -1 {
			FsmOnInitBetweenFloors()
		} else {
			FsmInit(ID)
			motorTimeoutStarted = timer.Get_wall_time()
		}
		fmt.Println("motorstop = ", elevator.MotorStop)
		time.Sleep(2 * time.Second)
	}
}

func RestartElevator() {
	outputDevice.MotorDirection(elev.MD_Stop)
	for i := 0; i < 300; i++ {
		time.Sleep(10 * time.Millisecond)
		outputDevice.MotorDirection(elev.MD_Stop)
	}
	fmt.Println("Starter heismotor på nytt, går videre")
	elevator.State = elev.IDLE

}

func ObstructionTimeout() {
	var prevObstructionState bool = false
	timeout_time := 8.0

	for {
		if elevator.Obs && (elevator.Obs != prevObstructionState) {
			ObstructionTimeoutStarted = timer.Get_wall_time()

		}
		if elevator.Obs && prevObstructionState && ((motorTimeoutStarted + timeout_time) < timer.Get_wall_time()) {
			fmt.Println("---------------------Obstruction timeout----------------------------")
			elevator.ObstructionFailure = true
			ObstructionTimeoutStarted = timer.Get_wall_time()
		}
		if !elevator.Obs {
			elevator.ObstructionFailure = false
		}

		prevObstructionState = elevator.Obs

		time.Sleep(200 * time.Millisecond)

	}

}
