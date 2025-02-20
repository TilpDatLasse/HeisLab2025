package main

type Elevator struct {
	floor    int
	dirn     int
	requests [N_FLOORS][N_BUTTONS]bool
	config   Config
}

type Config struct {
	clearRequestVariant int
}

type DirnBehaviourPair struct {
	dirn      int
	behaviour int
}

const (
	D_Up   = 1
	D_Down = -1
	D_Stop = 0

	EB_Moving   = 1
	EB_DoorOpen = 2
	EB_Idle     = 3

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

func requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.dirn {
	case D_Up:
		if requests_above(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{D_Down, EB_DoorOpen}
		} else if requests_below(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		}
	case D_Down:
		if requests_below(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{D_Up, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		}
	case D_Stop:
		if requests_here(e) {
			return DirnBehaviourPair{D_Stop, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requests_below(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		}
	}
	return DirnBehaviourPair{D_Stop, EB_Idle}
}

func requests_shouldStop(e Elevator) bool {
	switch e.dirn {
	case D_Down:
		return e.requests[e.floor][B_HallDown] || e.requests[e.floor][B_Cab] || !requests_below(e)
	case D_Up:
		return e.requests[e.floor][B_HallUp] || e.requests[e.floor][B_Cab] || !requests_above(e)
	case D_Stop:
	default:
		return true
	}
}

func requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type int) bool {
	switch e.config.clearRequestVariant {
	case CV_All:
		return e.floor == btn_floor
	case CV_InDirn:
		return e.floor == btn_floor &&
			(e.dirn == D_Up && btn_type == B_HallUp ||
				e.dirn == D_Down && btn_type == B_HallDown ||
				e.dirn == D_Stop ||
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
		case D_Up:
			if !requests_above(e) && !e.requests[e.floor][B_HallUp] {
				e.requests[e.floor][B_HallDown] = false
			}
			e.requests[e.floor][B_HallUp] = false
		case D_Down:
			if !requests_below(e) && !e.requests[e.floor][B_HallDown] {
				e.requests[e.floor][B_HallUp] = false
			}
			e.requests[e.floor][B_HallDown] = false
		case D_Stop:
		default:
			e.requests[e.floor][B_HallUp] = false
			e.requests[e.floor][B_HallDown] = false
		}
	}
	return e
}
