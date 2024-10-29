package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

const (
	checkpointDir  = "client_checkpoints"
	clientFilesDir = "client_files"
	blockSize      = 1048576 // 1 MB
)

func main() {
	fmt.Println("Client starting...")

	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		if os.Mkdir(checkpointDir, os.ModePerm) != nil {
			return
		}
	}

	files, err := os.ReadDir(clientFilesDir)
	if err != nil {
		fmt.Println("Error reading client files directory:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			sendFileInChunks(file.Name())
		}
	}

	fmt.Println("All files have been sent.")
}

func sendFileInChunks(filename string) {
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

	fmt.Println("Connected to server to send:", filename)

	checkpointPath := filepath.Join(checkpointDir, filename+".chk")
	startChunk := 0

	if chkFile, e := os.Open(checkpointPath); e == nil {
		scanner := bufio.NewScanner(chkFile)

		if scanner.Scan() {
			startChunk, _ = strconv.Atoi(scanner.Text())
		}

		if chkFile.Close() != nil {
			return
		}
	}

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

	_, err = fmt.Fprintf(conn, "%s:%d\n", filename, startChunk)
	if err != nil {
		fmt.Println("Error sending filename and checkpoint:", err)
	}

	_, err = file.Seek(int64(startChunk*blockSize), io.SeekStart)
	if err != nil {
		return
	}

	buffer := make([]byte, blockSize)
	chunkIndex := startChunk

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

		chunkIndex++
		updateCheckpoint(checkpointPath, chunkIndex)

		fmt.Println("Sent chuck:", chunkIndex)
	}

	fmt.Println("File sent successfully:", filename)
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
