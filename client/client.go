package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Client starting...")

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	fmt.Println("Connected to server")

	err = conn.Close()
	if err != nil {
		return
	}
}
