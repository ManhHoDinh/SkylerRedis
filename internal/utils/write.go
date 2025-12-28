package utils

import (
	"fmt"
	"net"
)

func WriteError(conn net.Conn, msg string) {
	conn.Write([]byte("-ERR " + msg + "\r\n"))
}

func WriteSimpleString(conn net.Conn, msg string) {
	conn.Write([]byte("+" + msg + "\r\n"))
}

func WriteBulkString(conn net.Conn, s string) {
	conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)))
}
func WriteArray(conn net.Conn, items []string) {
	conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(items))))
	for _, item := range items {
		WriteBulkString(conn, item)
	}
	fmt.Println("Wrote array to connection:", items)
}
func WriteInteger(conn net.Conn, n int) {
	conn.Write([]byte(fmt.Sprintf(":%d\r\n", n)))
}

func WriteNull(conn net.Conn) {
	conn.Write([]byte("$-1\r\n"))
}
