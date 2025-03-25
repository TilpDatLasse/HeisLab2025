package timer

import (
	"time"

	"github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
)

var (
	timer_channel = make(chan bool)
	timerEndTime  float64
	timerActive   bool
)

func Get_wall_time() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

func Timer_start(duration float64) {
	//go Time(timer_channel)
	timerEndTime = Get_wall_time() + duration
	timerActive = true
}

func Timer_stop() {
	timerActive = false
}

func Timer_timed_out() bool {
	return timerActive && Get_wall_time() > timerEndTime
}

func Time(reciever chan<- bool) {
	for {
		if Timer_timed_out() && !elevator_io.GetObstruction() {
			Timer_stop()
			reciever <- true
		}
		time.Sleep(15 * time.Millisecond)
	}
}
