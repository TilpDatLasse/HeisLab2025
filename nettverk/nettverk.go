package nettverk

import (
	"fmt"
	"os"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/localip"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"

)

var ID string
var InfoMap = make(map[string]InformationElev)
var WorldView [4][2]elev.ConfirmationState = [4][2]elev.ConfirmationState{{0, 0}, {0, 0}, {0, 0}, {0, 0}}
var HRArequest bool

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//	will be received as zero-values.

type HelloMsg struct {
	Message string
	Iter    int
}

type InformationElev struct {
	State        HRAElevState
	HallRequests [][2]elev.ConfirmationState // denne skal deles med alle peers, så alle vet hvilke ordre som er aktive
	ReadyForHRA  elev.ConfirmationState  // Når denne er !=0 skal ikke lenger info hentes fra elev-modulen
	ID           string
}

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

// henter status fra heisen og sender på channel som en informationElev-variabel
func SetElevatorStatus(ch_HRAInputTx chan InformationElev) {
	for {
		for HRArequest{
			// ingenting
			time.Sleep(10 * time.Millisecond)
		}
		info := Converter(fsm.FetchElevatorStatus())
		info.ID = ID
		ch_HRAInputTx <- info
		if HRArequest {
			info.ReadyForHRA = 1
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func BroadcastElevatorStatus(ch_HRAInputTx chan InformationElev) {
	for {
		b.Transmitter(14000, ch_HRAInputTx)
	}
}

func RecieveElevatorStatus(ch_HRAInputRx chan InformationElev) {
	for {
		b.Receiver(14000, ch_HRAInputRx)
	}
}


func Nettverk_hoved(ch_HRAInputRx chan InformationElev, id string) {

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	ID = id

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(16000, id, peerTxEnable)
	go peers.Receiver(16000, peerUpdateCh)

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-ch_HRAInputRx:  //heartbeat med info mottatt
			if a.ID != "" {
				InfoMap[a.ID] = a
				if a.ReadyForHRA != 0{  //hvis mottar at noen vil synke
					HRArequest = true // burde egt hente status fra egen heis før denne settes true
				}
			}
		// case a := <-ch_fromSync:
		// 	InfoMap = a
		}
	}
}

// konverterer en elev.elevator-variabel til en InformationElev-variabel
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

// konverterer states vi bruker til states HRA bruker
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

// konverterer dirn vi bruker til dirn HRA bruker
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

// Henter output fra HRA og sender videre til elev-modulen
func FromHRA(HRAOut chan map[string][][2]bool, ch_elevator_queue chan [][2]bool) {
	for {
		output := <-HRAOut
		for k, v := range output {
			if k == ID {
				ch_elevator_queue <- v
			}
		}
	}
}

func cabToBool(list []elev.ConfirmationState) []bool {
	boolList := make([]bool, len(list))
	for i, v := range list {
		boolList[i] = v != 0 // Convert non-zero values to true, zero to false
	}

	return boolList
}
