package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Print("Server starting...")

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			fmt.Println("Error closing listener:", err)
		}
	}(ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("Client connected:", conn.RemoteAddr())

		err = conn.Close()
		if err != nil {
			return
		}
	}
}
