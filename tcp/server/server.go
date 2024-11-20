// server.go
package main

import (
	"fmt"
	"net"
)

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Printf("Failed to start server: %v\n", err)
        return
    }
    defer listener.Close()
    fmt.Println("Server listening on :8080")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Failed to accept connection: %v\n", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 1024)

    for {
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Printf("Error reading: %v\n", err)
            return
        }

        // Echo back
        _, err = conn.Write(buffer[:n])
        if err != nil {
            fmt.Printf("Error writing: %v\n", err)
            return
        }
    }
}
