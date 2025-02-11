package main

import (
	"fmt"
)

type ElevatorBehaviour int

type Elevator struct {
	floor     int
	dirn      int
	behaviour ElevatorBehaviour
	requests  [N_FLOORS][N_BUTTONS]int
	config    ElevatorConfig
}

type ElevatorConfig struct {
	clearRequestVariant int
	doorOpenDurationS   float64
}

func ebToString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "EB_Idle"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	case EB_Moving:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"
	}
}

func elevatorPrint(es Elevator) {
	fmt.Println("  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		es.floor,
		elevioDirnToString(es.dirn),
		ebToString(es.behaviour),
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == B_HallUp) || (f == 0 && btn == B_HallDown) {
				fmt.Print("|     ")
			} else {
				if es.requests[f][btn] != 0 {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

func elevatorUninitialized() Elevator {
	return Elevator{
		floor:     -1,
		dirn:      D_Stop,
		behaviour: EB_Idle,
		config: ElevatorConfig{
			clearRequestVariant: CV_All,
			doorOpenDurationS:   3.0,
		},
	}
}
