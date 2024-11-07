package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	checkpointDir  = "server_checkpoints"
	serverFilesDir = "server_files"
	blockSize      = 1048576 // 1 MB
)

func main() {
	fmt.Println("Server starting...")

	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		if os.Mkdir(checkpointDir, os.ModePerm) != nil {
			return
		}
	}

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

	fmt.Println("Server is listening on port 8080...")

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

	header, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading filename and checkpoint:", err)
		return
	}

	parts := strings.Split(strings.TrimSpace(header), ":")
	filePath := filepath.Join(serverFilesDir, parts[0])

	startChunk, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("Error parsing chunk index:", err)
		return
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer func(file *os.File) {
		if file.Close() != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	_, err = file.Seek(int64(startChunk*blockSize), io.SeekStart)
	if err != nil {
		fmt.Println("Error seeking in file:", err)
		return
	}

	buffer := make([]byte, blockSize)
	chunkIndex := startChunk

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

		chunkIndex++
		updateCheckpoint(filepath.Join(checkpointDir, parts[0]+".chk"), chunkIndex)

		fmt.Println("Received chunk:", chunkIndex)
	}

	fmt.Println("File received and saved as:", filePath)
}

func updateCheckpoint(path string, chunkIndex int) {
	chkFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating checkpoint file:", err)
		return
	}

	defer func(chkFile *os.File) {
		if chkFile.Close() != nil {
			fmt.Println("Error closing checkpoint file:", err)
		}
	}(chkFile)

	_, err = chkFile.WriteString(strconv.Itoa(chunkIndex))
	if err != nil {
		fmt.Println("Error writing to checkpoint file:", err)
	}
}
