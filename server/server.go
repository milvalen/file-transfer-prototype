package main

import (
	"hash/crc32"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
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

func handleConnection(conn net.Conn, saveDir string) {
	defer conn.Close()

	var metadata FileMetadata

	// Получение метаданных
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&metadata); err != nil {
		fmt.Println("Ошибка при получении метаданных:", err)
		return
	}

	filePath := filepath.Join(saveDir, metadata.FileName)
	fmt.Printf("Получение файла: %s, размер: %d байт, CRC32: %s\n", metadata.FileName, metadata.FileSize, metadata.FileCRC)

	// Existing file check
	var currentSize int64
	if stat, err := os.Stat(filePath); err == nil {
		currentSize = stat.Size()
	}

	// TODO: refactor client response
	if currentSize <= metadata.FileSize {
		conn.Write([]byte("continue\n"))
		conn.Write([]byte(strconv.FormatInt(currentSize, 10) + "\n"))
	} else {
		conn.Write([]byte("new\n"))
		currentSize = 0
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()

	received := currentSize
	buffer := make([]byte, 1024*1024)
        i := 0

	for received < metadata.FileSize {
                frames := []string{"|", "/", "-", "\\"}
                i++
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Ошибка при получении данных:", err)
			return
		}
		_, err = file.Write(buffer[:n])
		if err != nil {
			fmt.Println("Ошибка при записи данных:", err)
			return
		}
		received += int64(n)
		fmt.Printf("\rПринято: %d/%d байт %s", received, metadata.FileSize, frames[i%3])
	}
	fmt.Println()

	calculatedCRC, err := calculateCRC32(filePath)
	if err != nil {
		fmt.Println("Ошибка при вычислении CRC32:", err)
		return
	}

	fmt.Printf("Ожидаемый CRC32: %s, полученный: %s\n", metadata.FileCRC, calculatedCRC)
	if calculatedCRC == metadata.FileCRC {
		conn.Write([]byte("CRC32_OK\n"))
		fmt.Println("Файл успешно передан и проверен!")
	} else {
		conn.Write([]byte("CRC32_ERROR\n"))
		fmt.Println("Ошибка проверки CRC32. Файл не соответствует!")
	}
}

func main() {
	port := 8080
	saveDir := "./files"

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		fmt.Println("Ошибка при создании директории:", err)
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Сервер запущен и ожидает подключения на порту %d, файлы сохраняются в %s\n", port, saveDir)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка подключения клиента:", err)
			continue
		}
		go handleConnection(conn, saveDir)
	}
}
