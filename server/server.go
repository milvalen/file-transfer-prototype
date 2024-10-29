package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const serverFilesDir = "server_files"

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

	_, err = io.Copy(file, reader)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Println("File received and saved as:", filename)
}
