package annet

type State int

const (
	INIT     State = 0
	IDLE           = 1
	MOVE           = 2
	STOP           = 3
	DOOROPEN       = 4
)

type Elevator struct {
	floor    int
	dirn     MotorDirection
	state    State
	requests [N_FLOORS][N_BUTTONS]bool
	config   ElevatorConfig
	obs      bool
}

type ElevatorConfig struct {
	clearRequestVariant int
	doorOpenDurationS   float64
}
