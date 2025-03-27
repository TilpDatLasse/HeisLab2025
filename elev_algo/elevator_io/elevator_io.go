package elevator_io

import (
	"fmt"
	"net"
	"sync"
	"time"
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

type ConfirmationState int

const (
	no_call      ConfirmationState = 0
	unregistered ConfirmationState = 1
	registered   ConfirmationState = 2
)

type State int

const (
	INIT     State = 0
	IDLE     State = 1
	MOVE     State = 2
	STOP     State = 3
	DOOROPEN State = 4
)

var (
	_initialized bool = false
	_numFloors   int  = 4
	_mtx         sync.Mutex
	_conn        net.Conn
	_pollRate    = 20 * time.Millisecond
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type Elevator struct {
	Floor       int
	Dirn        MotorDirection
	State       State
	Requests    [N_FLOORS][N_BUTTONS]ConfirmationState
	OwnRequests [N_FLOORS][N_BUTTONS]bool
	Config      ElevatorConfig
	Obs         bool
	MotorStop   bool
}

type ElevatorConfig struct {
	ClearRequestVariant int
	DoorOpenDurationS   float64
}

type ElevatorInputDevice struct {
	FloorSensor   func() int
	RequestButton func(ButtonType, int) bool
	stopButton    func() bool
	obstruction   func() bool
}

type ElevatorOutputDevice struct {
	FloorIndicator     func(int)
	RequestButtonLight func(ButtonType, int, ConfirmationState)
	DoorLight          func(bool)
	StopButtonLight    func(bool)
	MotorDirection     func(MotorDirection)
}

func Elevio_getInputDevice() ElevatorInputDevice {
	return ElevatorInputDevice{
		FloorSensor:   GetFloor,
		RequestButton: GetButton,
		stopButton:    GetStop,
		obstruction:   GetObstruction,
	}
}

func Elevio_getOutputDevice() ElevatorOutputDevice {
	return ElevatorOutputDevice{
		FloorIndicator:     SetFloorIndicator,
		RequestButtonLight: SetButtonLamp,
		DoorLight:          SetDoorOpenLamp,
		StopButtonLight:    SetStopLamp,
		MotorDirection:     SetMotorDirection,
	}
}

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}

	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetState(channel chan State) {
	channel <- INIT
}

func SetButtonLampsOff() {
	for f := 0; f < _numFloors; f++ {
		for b := ButtonType(0); b < 3; b++ {
			SetButtonLamp(b, f, 0)
		}
	}
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value ConfirmationState) {
	boolValue := false
	if value == 2 {
		boolValue = true
	}
	write([4]byte{2, byte(button), byte(floor), toByte(boolValue)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

// Update the order confirmationstates
func CyclicUpdate(list []ConfirmationState, wasTimedOut bool) ConfirmationState {
	isPresent := map[ConfirmationState]bool{} // stores the presence of each state
	for _, v := range list {
		isPresent[v] = true
	}
	switch {
	case isPresent[no_call] && isPresent[unregistered] && isPresent[registered]: //should ideally not happen
		fmt.Println("--------------2-1-0---------------")
		return unregistered // returns unregistered if it happens, so no orders are lost

	case !isPresent[no_call]: // Every peer has a unregistered or registered order, order should then be registered on all peers

		return registered
	case isPresent[registered] && isPresent[no_call]: // Order is served. No order should be returned
		if wasTimedOut { //checks if the information may be outdated in case of severe packetloss og network disconnect
			return registered
		} else {
			return no_call
		}

	case isPresent[no_call] && isPresent[unregistered]: // Some peer has a new order. Every peer should then get the order
		return unregistered

	case !isPresent[unregistered] && !isPresent[registered]:
		return no_call
	}
	return unregistered //default to loose no order
}
