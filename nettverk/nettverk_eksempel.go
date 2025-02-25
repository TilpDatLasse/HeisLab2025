package nettverk

import (
	"flag"
	"fmt"
	"os"
	"time"

	elev "github.com/TilpDatLasse/HeisLab2025/elev_algo/elevator_io"
	"github.com/TilpDatLasse/HeisLab2025/elev_algo/fsm"
	b "github.com/TilpDatLasse/HeisLab2025/nettverk/network/bcast"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/localip"
	"github.com/TilpDatLasse/HeisLab2025/nettverk/network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//
//	will be received as zero-values.

type ConfirmationState int

const (
	no_call      ConfirmationState = 0
	unregistered ConfirmationState = 1
	registered   ConfirmationState = 2
)

type HelloMsg struct {
	Message string
	Iter    int
}

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

// heihei

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
	confirmation [][2]ConfirmationState
}

func SetElevatorStatus(ch_HRAInputTx chan HRAInput) {
	for {
		ch_HRAInputTx <- Converter(fsm.FetchElevatorStatus())
		time.Sleep(1000 * time.Millisecond)
	}

}

func BroadcastElevatorStatus(ch_HRAInputTx chan HRAInput, statusElev chan elev.Elevator) {
	for {

		fmt.Println("yo1")
		b.Transmitter(14000, ch_HRAInputTx)
		fmt.Println("yo")

	}
}

func RecieveElevatorStatus(ch_HRAInputRx chan HRAInput) {
	for {
		fmt.Println("kjører recieved status")
		b.Receiver(14000, ch_HRAInputRx)
	}
}

func Nettverk_hoved(ch_HRAInputRx chan HRAInput) {

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(16000, id, peerTxEnable)
	go peers.Receiver(16000, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	//HelloTx := make(chan HelloMsg)

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	//go b.Transmitter(17000, bustatusElevttonTx)
	//go b.Receiver(17000, buttonRx)

	// The example message. We just send one of these every second.
	/*go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			HelloTx <- helloMsg
			time.Sleep(1 * time.Second)

		}BroadcastElevatorStatus
	}()*/

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		//case a := <-buttonRx:
		//buttonRx1 <- a
		//fmt.Printf("Received: %#v\n", a)
		case a := <-ch_HRAInputRx:
			fmt.Println("Reicievd status", a.States["one"].Floor)
		}
	}
}

func Converter(e elev.Elevator) HRAInput {
	fmt.Println("converting")
	rawInput := e
	hallRequests := make([][2]bool, len(rawInput.Requests))
	cabRequests := make([]bool, len(rawInput.Requests))

	for i := 0; i < len(rawInput.Requests); i++ {
		hallRequests[i] = [2]bool{rawInput.Requests[i][0], rawInput.Requests[i][1]}
		cabRequests[i] = rawInput.Requests[i][2]
	}

	fmt.Println("raw:", hallRequests[1][1])

	input := HRAInput{
		HallRequests: hallRequests,
		States: map[string]HRAElevState{
			"one": HRAElevState{
				Behavior:    string(rawInput.State),
				Floor:       rawInput.Floor,
				Direction:   string(rawInput.Dirn),
				CabRequests: cabRequests,
			},
		},
	}

	return input

}
