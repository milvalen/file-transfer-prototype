package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const clientFilesDir = "client_files"

func main() {
	fmt.Println("Client starting...")

	files, err := os.ReadDir(clientFilesDir)
	if err != nil {
		fmt.Println("Error reading client files directory:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			sendFile(file.Name())
		}
	}
}

func sendFile(filename string) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	defer func(conn net.Conn) {
		if conn.Close() != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	fmt.Println("Connected to server to send", filename)

	file, err := os.Open(filepath.Join(clientFilesDir, filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer func(file *os.File) {
		if file.Close() != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	_, err = fmt.Fprintln(conn, filename)
	if err != nil {
		fmt.Println("Error sending filename:", err)
	}

	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file:", err)
		return
	}

	fmt.Println("File sent successfully:", filename)
}
