package worldview

import (
	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
)

// Converting an elev.Elevator-variabel to an InformationElev-variabel
func Converter(e elev.Elevator) InformationElev {
	rawInput := e
	hallRequests := make([][2]elev.ConfirmationState, len(rawInput.Requests))
	cabRequests := make([]elev.ConfirmationState, len(rawInput.Requests))

	for i := 0; i < len(rawInput.Requests); i++ {
		hallRequests[i] = [2]elev.ConfirmationState{rawInput.Requests[i][0], rawInput.Requests[i][1]}
		cabRequests[i] = rawInput.Requests[i][2]
	}

	input := InformationElev{
		HallRequests: hallRequests,
		State: HRAElevState{
			Behavior:    stateToString(rawInput.State),
			Floor:       rawInput.Floor,
			Direction:   dirnToString(rawInput.Dirn),
			CabRequests: cabToBool(cabRequests),
		},
	}
	return input
}

// Converting elev.state to HRAElevState.Behaviour (string)
func stateToString(s elev.State) string {
	switch s {
	case elev.IDLE:
		return "idle"
	case elev.MOVE:
		return "moving"
	case elev.DOOROPEN:
		return "doorOpen"
	case elev.STOP:
		return "doorOpen"
	default:
		return "idle"
	}
}

// Converting elev.MotorDirection to HRAElevState.Direction
func dirnToString(s elev.MotorDirection) string {
	switch s {
	case elev.MD_Up:
		return "up"
	case elev.MD_Down:
		return "down"
	case elev.MD_Stop:
		return "stop"
	default:
		return "stop"
	}
}

// Converting cabrequests from ConfirmationState (cyclic-counter) to bool
// Notice that cab-requests do not need to be confirmed, so both confirmationstates 1 and 2 will yield true
func cabToBool(list []elev.ConfirmationState) []bool {
	boolList := make([]bool, len(list))
	for i, v := range list {
		boolList[i] = v != 0
	}
	return boolList
}

func HallToBool(hallReqList [][2]elev.ConfirmationState) [][2]bool {
	boolList := make([][2]bool, len(hallReqList))
	for i, v := range hallReqList {
		boolList[i][0] = v[0] == 2
		boolList[i][1] = v[1] == 2
	}
	return boolList
}