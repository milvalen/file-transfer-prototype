package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const (
	serverFilesDir = "server_files"
	blockSize      = 1048576 // 1 MB
)

func main() {
	fmt.Print("Server starting...")

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	defer func(ln net.Listener) {
		if ln.Close() != nil {
			fmt.Println("Error closing listener:", err)
		}
	}(ln)

	for {
		conn, e := ln.Accept()
		if e != nil {
			fmt.Println("Error accepting connection:", e)
			continue
		}

		fmt.Println("Client connected:", conn.RemoteAddr())

		go handleFileTransfer(conn)
	}
}

func handleFileTransfer(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	reader := bufio.NewReader(conn)

	filename, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading filename:", err)
		return
	}

	filename = filename[:len(filename)-1]
	filePath := filepath.Join(serverFilesDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer func(file *os.File) {
		if file.Close() != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	buffer := make([]byte, blockSize)

	for {
		bytesRead, e := reader.Read(buffer)
		if e != nil {
			if e == io.EOF {
				break
			}

			fmt.Println("Error reading chunk:", e)
			return
		}

		_, err = file.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Println("Error writing chunk to file:", err)
			return
		}

		fmt.Println("Received chunk of size:", bytesRead)
	}

	fmt.Println("File received and saved as:", filename)
}
