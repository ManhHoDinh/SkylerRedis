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
			// client closed connection
			return
		}

		// Basic RESP parsing: wait for "PING"
		if strings.TrimSpace(line) == "*1" {
			// Read the next 2 lines: "$4" and "PING"
			_, _ = reader.ReadString('\n') // skip "$4"
			cmd, _ := reader.ReadString('\n')
			cmd = strings.TrimSpace(cmd)

			if strings.ToUpper(cmd) == "PING" {
				conn.Write([]byte("+PONG\r\n"))
			}
		}
	}
}
