package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const (
	clientFilesDir = "client_files"
	blockSize      = 1048576 // 1 MB
)

func main() {
	fmt.Println("Client starting...")

	files, err := os.ReadDir(clientFilesDir)
	if err != nil {
		fmt.Println("Error reading client files directory:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			sendFileInChuncks(file.Name())
		}
	}
}

func sendFileInChuncks(filename string) {
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

	buffer := make([]byte, blockSize)

	for {
		bytesRead, e := file.Read(buffer)
		if e != nil && e != io.EOF {
			fmt.Println("Error reading file:", e)
			return
		}

		if bytesRead == 0 {
			break
		}

		_, err = conn.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Println("Error sending chunk:", err)
			return
		}

		fmt.Println("Sent chunk of size:", bytesRead)
	}

	fmt.Println("File sent successfully:", filename)
}
