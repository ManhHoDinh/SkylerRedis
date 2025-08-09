package main

import (
	"SkylerRedis/app/handler"
	"SkylerRedis/app/server"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var port = flag.String("port", "6379", "Port for redis server")
var replicaof = flag.String("replicaof", "", "Address of the master server")
var Server server.Server
func main() {
	flag.Parse() // Parse command line arguments
	l, err := net.Listen("tcp", "0.0.0.0:"+*port)
	if err != nil {
		fmt.Println("Failed to bind to port", *port)
		os.Exit(1)
	}
	Server = server.Server{
		Port:      *port,
	}
	if *replicaof != "" {
		Server.IsMaster = false
		fmt.Println("Server is running on port", *port)
		fmt.Println("Replica of:", *replicaof)

	} else {
		Server.IsMaster = true
		fmt.Println("Running as master")
	}


	go func() {
		for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handler.HandleConnection(conn, Server)
	}
	}()
	if *replicaof != "" {
		parts := strings.Split(*replicaof, " ")
		masterAddr := parts[0] + ":" + parts[1]

		go func() {
        for {
            conn, err := net.Dial("tcp", masterAddr)
            if err != nil {
                fmt.Println("Failed to connect to master:", err)
                time.Sleep(5 * time.Second)
                continue
            }

            fmt.Println("Connected to master:", masterAddr)
            for {
                _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
                if err != nil {
                    fmt.Println("Connection to master lost:", err)
                    conn.Close()
                    break
                }
                fmt.Println("Sent PING to master")
                time.Sleep(5 * time.Second)
            }
        }
    }()
	}
	select {} // Keep the main goroutine running
}