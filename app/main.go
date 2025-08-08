package main

import (
	"SkylerRedis/app/handler"
	"flag"
	"fmt"
	"net"
	"os"
)

var port = flag.String("port", "6379", "Port for redis server")

func main() {
	flag.Parse() // Parse command line arguments
	l, err := net.Listen("tcp", "0.0.0.0:"+*port)
	if err != nil {
		fmt.Println("Failed to bind to port", *port)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		go handler.HandleConnection(conn)
	}
}
