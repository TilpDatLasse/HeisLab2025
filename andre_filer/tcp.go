package main

/*
import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Define the server address and port
	address := "10.100.23.204:34933" // Replace with the target IP and port

	// Connect to the server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close() // Ensure the connection is closed when done

	fmt.Println("Connected to server:", address)

	go send(conn)
	go recieve(conn)
	time.Sleep(20 * time.Second)
}

func send(conn net.Conn) {
	// Send data to the server_, err := conn.Read(buffer)
	for i := 0; i < 4; i++ {
		message := "Hello, Lasse\n\x00"
		_, err := conn.Write([]byte(message + "\n")) // Add newline for easier reading by the server
		if err != nil {
			fmt.Println("Error sending data:", err)
			return
		}
		fmt.Println("Sent:", message)
		time.Sleep(2 * time.Second)
	}
}

func recieve(conn net.Conn) {
	// Receive data from the server
	buffer := make([]byte, 1024)
	for i := 0; i < 4; i++ {

		//Read response from the server
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}
		fmt.Printf("Received: %s", string(buffer))
	}
}
*/
