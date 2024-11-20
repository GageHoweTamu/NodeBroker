// client.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter message: ")
		scanner.Scan()
		message := scanner.Text()

		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Error writing: %v\n", err)
			return
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading: %v\n", err)
			return
		}

		fmt.Printf("Server response: %s\n", buffer[:n])
	}
}
