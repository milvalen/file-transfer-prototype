package main

import (
	"hash/crc32"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

type FileMetadata struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileCRC  string `json:"file_crc"`
}

func calculateCRC32(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%08x", hash.Sum32()), nil
}

func sendFile(serverIP string, port int, filePath string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fileCRC, err := calculateCRC32(filePath)
	if err != nil {
		return err
	}

	metadata := FileMetadata{
		FileName: filepath.Base(filePath),
		FileSize: fileInfo.Size(),
		FileCRC:  fileCRC,
	}

	// Metadata send
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&metadata); err != nil {
		return err
	}

	// TODO: debug server response
	var response string
	fmt.Fscanln(conn, &response)
	var currentSize int64
	if response == "continue" {
		fmt.Fscanln(conn, &currentSize)
		fmt.Printf("Продолжаем с позиции %d байт\n", currentSize)
	} else {
		currentSize = 0
		fmt.Println("Начинаем новую передачу.")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Seek(currentSize, io.SeekStart)
	buffer := make([]byte, 1024*1024)
	sent := currentSize
        i := 0

	for {
                frames := []string{"|", "/", "-", "\\"}
                i++
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			return err
		}
		sent += int64(n)
		fmt.Printf("\rОтправлено: %d/%d байт %s", sent, metadata.FileSize, frames[i%3])
	}
	fmt.Println()

	fmt.Println("Ожидание проверки CRC32...")
	fmt.Fscanln(conn, &response)
	if response == "CRC32_OK" {
		fmt.Println("Передача завершена успешно. CRC32 совпадает.")
	} else {
		fmt.Println("Ошибка: CRC32 не совпадает.")
	}
	return nil
}

func main() {
	var serverIP string
	var filePath string

	fmt.Print("Введите IP-адрес сервера: ")
	fmt.Scanln(&serverIP)

	fmt.Print("Введите путь к файлу: ")
	fmt.Scanln(&filePath)

	if err := sendFile(serverIP, 8080, filePath); err != nil {
		fmt.Println("Ошибка:", err)
	}
}
