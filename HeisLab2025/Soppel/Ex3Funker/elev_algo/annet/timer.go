package annet

import (
	"time"
)

func get_wall_time() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

var (
	timerEndTime float64
	timerActive  bool
)

func timer_start(duration float64) {
	timerEndTime = get_wall_time() + duration
	timerActive = true
}

func Timer_stop() {
	timerActive = false
}

func Timer_timed_out() bool {
	return timerActive && get_wall_time() > timerEndTime
}
