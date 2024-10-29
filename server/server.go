package main

import (
	"bufio"
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

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	fmt.Println("Received message from client:", message)

	_, err = fmt.Fprintf(conn, "Message received")
	if err != nil {
		return
	}
}
