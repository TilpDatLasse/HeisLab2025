package annet

type Behaviour int

const (
	EB_Moving    Behaviour = 1
	EB_DoorOpen            = 2
	EB_Idle                = 3
	EB_UNDEFINED           = 0
)

type Elevator struct {
	floor     int
	dirn      MotorDirection
	behaviour Behaviour
	requests  [N_FLOORS][N_BUTTONS]bool
	config    ElevatorConfig
}

type ElevatorConfig struct {
	clearRequestVariant int
	doorOpenDurationS   float64
}
