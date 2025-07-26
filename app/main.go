package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var store = make(map[string]string)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Đọc dòng đầu tiên như: *3
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "*") {
			conn.Write([]byte("-ERR invalid format\r\n"))
			continue
		}
		numArgs := parseLength(line)

		args := []string{}
		for i := 0; i < numArgs; i++ {
			_, err := reader.ReadString('\n') // Skip $len
			if err != nil {
				return
			}
			arg, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			args = append(args, strings.TrimSpace(arg))
		}

		if len(args) == 0 {
			conn.Write([]byte("-ERR empty command\r\n"))
			continue
		}

		switch strings.ToUpper(args[0]) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'ECHO'\r\n"))
			} else {
				conn.Write([]byte("+" + args[1] + "\r\n"))
			}
		case "SET":
			if len(args) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SET'\r\n"))
			} else {
				store[args[1]] = args[2]
				conn.Write([]byte("+OK\r\n"))
			}
		case "GET":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'GET'\r\n"))
			} else {
				val, ok := store[args[1]]
				if ok {
					conn.Write([]byte("+" + val + "\r\n"))
				} else {
					conn.Write([]byte("$-1\r\n")) // nil
				}
			}
		default:
			conn.Write([]byte("-ERR unknown command '" + args[0] + "'\r\n"))
		}
	}
}

func parseLength(s string) int {
	var n int
	fmt.Sscanf(s, "*%d", &n)
	return n
}
