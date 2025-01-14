package main

import (
	"fmt"
	"net"
	//"strings"
)

func main() {
	// sendAddr := "10.100.23.204:20015"

	// Bind socket2 to local address
	

	socket2, err := net.ListenPacket("udp", "0.0.0.0:20015")
	if err != nil {
		fmt.Println("Could not bind on socket 2:", err)
		return
	}
	defer socket2.Close()

	// Bind socket1 to local address
	socket1, err := net.ListenPacket("udp", "0.0.0.0:30000")
	if err != nil {
		fmt.Println("Could not bind on socket 1:", err)
		return
	}
	defer socket1.Close()

	buffer := make([]byte, 1024)
	receivedData := false

	// Receiving data on socket1
	for !receivedData {
		n, srcAddr, err := socket1.ReadFrom(buffer)
		if err != nil {
			fmt.Println("Failed to receive from socket 1:", err)
			return
		}
		if n > 0 {
			receivedData = true
			fmt.Printf("Received %d bytes from %s: %s\n", n, srcAddr.String(), string(buffer[:n]))
		}
	}

	// Sending data on socket2
	sendData := "Hello from group 36\n"
	_, err = socket2.WriteTo([]byte(sendData), &net.UDPAddr{IP: net.ParseIP("10.100.23.204"), Port: 20015})
	if err != nil {
		fmt.Println("Error sending to socket 2:", err)
		return
	}

	receivedData = false

	// Receiving data on socket2
	for !receivedData {
		n, srcAddr, err := socket2.ReadFrom(buffer)
		if err != nil {
			fmt.Println("Failed to receive from socket 2:", err)
			return
		}
		if n > 0 {
			fmt.Println("antall n ", n)
			receivedData = true
			fmt.Printf("Received %d bytes from %s: %s\n", n, srcAddr.String(), string(buffer[:n]))
		}
	}
}
