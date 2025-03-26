package requests

import (
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
)

func RequestsAbove(e elev.Elevator) bool {
	for f := e.Floor + 1; f < elev.N_FLOORS; f++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			if e.OwnRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elev.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			if e.OwnRequests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elev.Elevator) bool {
	for btn := 0; btn < elev.N_BUTTONS; btn++ {
		if e.OwnRequests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirection(e elev.Elevator) (elev.MotorDirection, elev.State) {
	switch e.Dirn {
	case elev.MD_Up:
		if RequestsAbove(e) {
			return elev.MD_Up, elev.MOVE
		} else if requestsHere(e) {
			return elev.MD_Down, elev.DOOROPEN
		} else if requestsBelow(e) {
			return elev.MD_Down, elev.MOVE
		}
	case elev.MD_Down:
		if requestsBelow(e) {
			return elev.MD_Down, elev.MOVE
		} else if requestsHere(e) {
			return elev.MD_Up, elev.DOOROPEN
		} else if RequestsAbove(e) {
			return elev.MD_Up, elev.MOVE
		}
	case elev.MD_Stop:
		if requestsHere(e) {
			return elev.MD_Stop, elev.DOOROPEN
		} else if RequestsAbove(e) {
			return elev.MD_Up, elev.MOVE
		} else if requestsBelow(e) {
			return elev.MD_Down, elev.MOVE
		}
	}
	return elev.MD_Stop, elev.IDLE
}

func ShouldStop(e elev.Elevator) bool {
	switch e.Dirn {
	case elev.MD_Down:
		return e.OwnRequests[e.Floor][elev.B_HallDown] || e.OwnRequests[e.Floor][elev.B_Cab] || !requestsBelow(e)
	case elev.MD_Up:
		return e.OwnRequests[e.Floor][elev.B_HallUp] || e.OwnRequests[e.Floor][elev.B_Cab] || !RequestsAbove(e)
	default:
		return true
	}

}

func ShouldClearImmediately(e elev.Elevator, btn_floor int, btn_type int) bool {
	switch e.Config.ClearRequestVariant {
	case elev.CV_All:
		return e.Floor == btn_floor
	case elev.CV_InDirn:
		return e.Floor == btn_floor &&
			(e.Dirn == elev.MD_Up && btn_type == elev.B_HallUp ||
				e.Dirn == elev.MD_Down && btn_type == elev.B_HallDown ||
				e.Dirn == elev.MD_Stop ||
				btn_type == elev.B_Cab)
	default:
		return false
	}
}

func ClearAtCurrentFloor(e elev.Elevator) elev.Elevator {
	switch e.Config.ClearRequestVariant {
	case elev.CV_All:
		for btn := 0; btn < elev.N_BUTTONS; btn++ {
			e.OwnRequests[e.Floor][btn] = false
		}
	case elev.CV_InDirn:
		e.OwnRequests[e.Floor][elev.B_Cab] = false
		e.Requests[e.Floor][elev.B_Cab] = 0
		switch e.Dirn {
		case elev.MD_Up:
			if !RequestsAbove(e) && !e.OwnRequests[e.Floor][elev.B_HallUp] {
				e.OwnRequests[e.Floor][elev.B_HallDown] = false
				e.Requests[e.Floor][elev.B_HallDown] = 0
			}
			e.OwnRequests[e.Floor][elev.B_HallUp] = false
			e.Requests[e.Floor][elev.B_HallUp] = 0

		case elev.MD_Down:
			if !requestsBelow(e) && !e.OwnRequests[e.Floor][elev.B_HallDown] {
				e.OwnRequests[e.Floor][elev.B_HallUp] = false
				e.Requests[e.Floor][elev.B_HallUp] = 0
			}
			e.OwnRequests[e.Floor][elev.B_HallDown] = false
			e.Requests[e.Floor][elev.B_HallDown] = 0
		case elev.MD_Stop:
			e.OwnRequests[e.Floor][elev.B_HallUp] = false
			e.Requests[e.Floor][elev.B_HallUp] = 0
			e.OwnRequests[e.Floor][elev.B_HallDown] = false
			e.Requests[e.Floor][elev.B_HallDown] = 0

		default:
			e.OwnRequests[e.Floor][elev.B_HallUp] = false
			e.Requests[e.Floor][elev.B_HallUp] = 0
			e.OwnRequests[e.Floor][elev.B_HallDown] = false
			e.Requests[e.Floor][elev.B_HallDown] = 0
		}
	}
	return e
}
