package utils

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
)

func ParseArgs(conn net.Conn, reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" || !strings.HasPrefix(line, "*") {
		WriteError(conn, "invalid format")
		return nil, err
	}
	n := parseLength(line)
	args := []string{}
	for i := 0; i < n; i++ {
		_, err = reader.ReadString('\n') // skip $len
		if err != nil {
			WriteError(conn, err.Error())
			return nil, err
		}
		arg, err := reader.ReadString('\n')
		if err != nil {
			WriteError(conn, err.Error())
			return nil, err
		}
		args = append(args, strings.TrimSpace(arg))
	}
	if len(args) == 0 {
		WriteError(conn, "empty command")
	}
	return args, nil
}

func parseLength(s string) int {
	var n int
	fmt.Sscanf(s, "*%d", &n)
	return n
}

// ParseArgsWithByteCount parses RESP command and returns both args and byte count
func ParseArgsWithByteCount(conn net.Conn, reader *bufio.Reader) ([]string, int, error) {
	var totalBytes int

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, 0, err
	}
	totalBytes += len(line)

	line = strings.TrimSpace(line)
	if line == "" || !strings.HasPrefix(line, "*") {
		WriteError(conn, "invalid format")
		return nil, 0, errors.New("invalid format")
	}

	n := parseLength(line)
	args := []string{}

	for i := 0; i < n; i++ {
		lenLine, err := reader.ReadString('\n') // read $len line
		if err != nil {
			return nil, 0, err
		}
		totalBytes += len(lenLine)

		arg, err := reader.ReadString('\n') // read actual argument
		if err != nil {
			return nil, 0, err
		}
		totalBytes += len(arg)

		args = append(args, strings.TrimSpace(arg))
	}

	return args, totalBytes, nil
}
func FormatAddr(host, port string) string {
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		return fmt.Sprintf("[%s]:%s", host, port) // IPv6
	}
	return fmt.Sprintf("%s:%s", host, port) // IPv4 / hostname
}
func ArgsToRESP(args []string) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, arg := range args {
		b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	return []byte(b.String())
}

func IsModifyCommand(args []string) bool {
	command := strings.ToUpper(args[0])
	return len(args) > 0 && (command == "SET" || command == "DEL" || command == "LPUSH" || command == "RPUSH" || command == "INCR")
}
