package annet

import (
	"fmt"
	"time"
)

var timer_channel = make(chan bool)

func get_wall_time() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

var (
	timerEndTime float64
	timerActive  bool
)

func timer_start(duration float64) {
	fmt.Println("timer started")
	go Time(timer_channel)
	timerEndTime = get_wall_time() + duration
	timerActive = true
}

func Timer_stop() {
	timerActive = false
}

func Timer_timed_out() bool {

	return timerActive && get_wall_time() > timerEndTime

}

func Time(reciever chan<- bool) {
	for {

		if Timer_timed_out() && !elevator.obs {
			reciever <- true
		}

	}

}
