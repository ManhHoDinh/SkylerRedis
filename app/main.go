package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var store = make(map[string]Entry)
var rPlush = make(map[string][]string)
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
			if len(args) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SET'\r\n"))
				break
			}

			key := args[1]
			val := args[2]
			var expiry time.Time

			// parse optional arguments (PX ...)
			if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
				ms, err := strconv.Atoi(args[4])
				if err != nil {
					conn.Write([]byte("-ERR PX value must be integer\r\n"))
					break
				}
				expiry = time.Now().Add(time.Duration(ms) * time.Millisecond)
			}

			store[key] = Entry{value: val, expiryTime: expiry}
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'GET'\r\n"))
			} else {
				if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'GET'\r\n"))
				break
				}
				key := args[1]
				entry, ok := store[key]
				if !ok || (entry.expiryTime != (time.Time{}) && time.Now().After(entry.expiryTime)) {
					delete(store, key) 
					conn.Write([]byte("$-1\r\n"))
				} else {
					conn.Write([]byte("+" + entry.value + "\r\n"))
				}
			}
		case "RPUSH":
			if len(args) < 3{
				conn.Write([]byte("-ERR wrong number of arguments for 'RPUSH'\r\n"))
			} else {
				for i := 2; i < len(args); i++ {
					rPlush[args[1]] = append(rPlush[args[1]], args[i])
				}
				conn.Write([]byte(":" + strconv.Itoa(len(rPlush[args[1]])) + "\r\n"))
			}
		case "LRANGE":
			if len(args) != 4 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LRANGE'\r\n"))
				break
			}

			start, err1 := strconv.Atoi(args[2])
			end, err2 := strconv.Atoi(args[3])
			
			if err1 != nil || err2 != nil {
				conn.Write([]byte("-ERR invalid start or end index\r\n"))
				break
			}
			list, exist := rPlush[args[1]]
			
			if start < 0 {
				start = len(list) + start
				if start < 0 {
					start = 0
				}	
			}
			if end < 0 {
				end = len(list) + end
				if end < 0 {
					end = 0
				}
			}

			if  !exist || start >= len(list) || start > end {
				conn.Write([]byte("*0\r\n"))
				break
			}

			if end >= len(list) {
				// Case 3
				end = len(list) - 1
			}

			sublist := list[start : end+1]

			// RESP encode result
			conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(sublist))))
			for _, item := range sublist {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(item), item)))
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
type Entry struct {
	value string
	expiryTime time.Time
} 