package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)

		switch line {
		case "*1":
			cmd, err := getCommand(reader)
			if err != nil {
				return
			}
			if strings.ToUpper(cmd) == "PING" {
				conn.Write([]byte("+PONG\r\n"))
			} else {
				conn.Write([]byte("-ERR unknown command '" + cmd + "'\r\n"))
			}
		case "*2":
			cmd, err := getCommand(reader)
			if err != nil {
				return
			}
			if strings.ToUpper(cmd) == "ECHO" {
				msg, err := getCommand(reader)
				if err != nil {
					return
				}
				conn.Write([]byte("+" + msg + "\r\n"))
			} else {
				conn.Write([]byte("-ERR unknown command '" + cmd + "'\r\n"))
			}
		default:
			conn.Write([]byte("-ERR unknown format\r\n"))
		}
	}
}

func getCommand(reader *bufio.Reader) (string, error) {
	// Skip the length line, e.g., "$4"
	_, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Read the actual command or argument
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
