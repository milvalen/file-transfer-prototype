package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Client starting...")

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	fmt.Println("Connected to server")
	fmt.Print("Enter message: ")

	message, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	_, err = fmt.Fprint(conn, message)
	if err != nil {
		return
	}

	reply, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("Server reply: ", reply)
}
