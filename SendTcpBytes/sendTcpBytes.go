package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func sendBytesToTarget(data []byte, targetAddress string) error {
	// Establish a TCP connection to the target address
	conn, err := net.Dial("tcp", targetAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to target: %w", err)
	}
	defer conn.Close()

	// Send the byte array to the target address
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	fmt.Printf("Data sent successfully to %s\n", targetAddress)
	return nil
}

func main() {
	// Example byte array
	data := []byte{
		69, 76, 80, 0, 0, 2, 0, 1, 7, 107, 0, 0, 7, 213, 0, 64, 196, 203, 107, 41,
		72, 240, 0, 50, 0, 93, 0, 2, 128, 3, 4, 33, 184, 248, 83, 202, 23, 34,
		255, 191, 0, 1, 0, 0, 0, 0, 7, 213, 38, 148, 203, 144, 24, 43, 255, 181,
		0, 11, 0, 0, 0, 0, 7, 213, 36, 148, 203, 144, 24, 107, 255, 181, 0, 11,
		0, 0, 0, 0, 7, 213,
	}

	// Target address (replace with actual address and port)
	targetAddress := "172.31.12.230:8552"

	fmt.Println("Press Enter to send the message. Type 'q' to quit.")

	// Create a buffered reader for user input
	reader := bufio.NewReader(os.Stdin)

	for {
		// Wait for the user to press Enter
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Check if the user wants to exit
		if input == "q\n" {
			fmt.Println("Exiting...")
			break
		}

		// Send the bytes to the target
		if err := sendBytesToTarget(data, targetAddress); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

