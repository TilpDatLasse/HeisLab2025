package annet

import (
	"fmt"
)

const (
	B_HallUp   = 0
	B_HallDown = 1
	B_Cab      = 2

	CV_All    = 0
	CV_InDirn = 1

	N_FLOORS  = 4
	N_BUTTONS = 3
)

func requests_above(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func requests_chooseDirection(e Elevator) (MotorDirection, Behaviour) {
	fmt.Println("chooseDir")
	switch e.dirn {
	case MD_Up:
		if requests_above(e) {
			return MD_Up, EB_Moving
		} else if requests_here(e) {
			return MD_Down, EB_DoorOpen
		} else if requests_below(e) {
			return MD_Down, EB_Moving
		}
	case MD_Down:
		if requests_below(e) {
			return MD_Down, EB_Moving
		} else if requests_here(e) {
			return MD_Up, EB_DoorOpen
		} else if requests_above(e) {
			return MD_Up, EB_Moving
		}
	case MD_Stop:
		if requests_here(e) {
			return MD_Stop, EB_DoorOpen
		} else if requests_above(e) {
			return MD_Up, EB_Moving
		} else if requests_below(e) {
			return MD_Down, EB_Moving
		}
	}
	return MD_Stop, EB_Idle
}

func requests_shouldStop(e Elevator) bool {
	switch e.dirn {
	case MD_Down:
		return e.requests[e.floor][B_HallDown] || e.requests[e.floor][B_Cab] || !requests_below(e)
	case MD_Up:
		return e.requests[e.floor][B_HallUp] || e.requests[e.floor][B_Cab] || !requests_above(e)
	case MD_Stop:
	default:
		return true
	}
	return true
}

func requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type int) bool {
	switch e.config.clearRequestVariant {
	case CV_All:
		return e.floor == btn_floor
	case CV_InDirn:
		return e.floor == btn_floor &&
			(e.dirn == MD_Up && btn_type == B_HallUp ||
				e.dirn == MD_Down && btn_type == B_HallDown ||
				e.dirn == MD_Stop ||
				btn_type == B_Cab)
	default:
		return false
	}
}

func requests_clearAtCurrentFloor(e Elevator) Elevator {
	switch e.config.clearRequestVariant {
	case CV_All:
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.requests[e.floor][btn] = false
		}
	case CV_InDirn:
		e.requests[e.floor][B_Cab] = false
		switch e.dirn {
		case MD_Up:
			if !requests_above(e) && !e.requests[e.floor][B_HallUp] {
				e.requests[e.floor][B_HallDown] = false
			}
			e.requests[e.floor][B_HallUp] = false
		case MD_Down:
			if !requests_below(e) && !e.requests[e.floor][B_HallDown] {
				e.requests[e.floor][B_HallUp] = false
			}
			e.requests[e.floor][B_HallDown] = false
		case MD_Stop:
		default:
			e.requests[e.floor][B_HallUp] = false
			e.requests[e.floor][B_HallDown] = false
		}
	}
	return e
}
