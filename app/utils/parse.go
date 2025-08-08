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
		return nil, errors.New("invalid format")
	}
	n := parseLength(line)
	args := []string{}
	for i := 0; i < n; i++ {
		_, err = reader.ReadString('\n') // skip $len
		if err != nil {
			return nil, err
		}
		arg, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		args = append(args, strings.TrimSpace(arg))
	}
	return args, nil
}

func parseLength(s string) int {
	var n int
	fmt.Sscanf(s, "*%d", &n)
	return n
}
