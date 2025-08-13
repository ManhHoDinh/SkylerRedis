package main

import (
	"SkylerRedis/app/handler"
	"SkylerRedis/app/server"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var port = flag.String("port", "6379", "Port for redis server")
var replicaof = flag.String("replicaof", "", "Address of the master server")
var hostname = flag.String("hostname", "0.0.0.0", "Hostname for the server")

func main() {
	flag.Parse() // Parse command line arguments
	l, err := net.Listen("tcp", *hostname+":"+*port)
	if err != nil {
		fmt.Println("Failed to bind to port", *port)
		os.Exit(1)
	}
	Server := server.Server{}
	if *replicaof != "" {
		fmt.Println("Server is running on port", *port)
		fmt.Println("Replica of:", *replicaof)

	} else {
		fmt.Println("Running as master")
		Server.IsMaster = true
	}
	go sendToMaster()
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Failed to accept connection:", err)
				continue
			}
			Server.Conn = conn
			go handler.HandleConnection(Server)
		}
	}()
	select {} // Keep the main goroutine running
}
func sendToMaster() {
	if *replicaof != "" {
		parts := strings.Split(*replicaof, " ")
		masterAddr := parts[0] + ":" + parts[1]
		conn, err := net.Dial("tcp", masterAddr)
		if err != nil {
			fmt.Println("Failed to connect to master:", err)
			return
		}

		fmt.Println("Connected to master:", masterAddr)

		// Start goroutine to handle master communication and command propagation
		go func() {
			defer conn.Close()

			// Send PING
			_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
			if err != nil {
				fmt.Println("Connection to master lost:", err)
				return
			}
			fmt.Println("Sent PING to master")

			// Read PONG response
			buffer := make([]byte, 1024)
			_, err = conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading PONG:", err)
				return
			}

			// Send first REPLCONF
			_, err = conn.Write([]byte(fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%s\r\n", *port)))
			if err != nil {
				fmt.Println("Connection to master lost:", err)
				return
			}
			fmt.Println("Sent first REPLCONF to master")

			// Read OK response
			_, err = conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading first REPLCONF OK:", err)
				return
			}

			// Send second REPLCONF
			_, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
			if err != nil {
				fmt.Println("Connection to master lost:", err)
				return
			}
			fmt.Println("Sent Second REPLCONF to master")

			// Read OK response
			_, err = conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading second REPLCONF OK:", err)
				return
			}

			// Send PSYNC
			_, err = conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
			if err != nil {
				fmt.Println("Connection to master lost:", err)
				return
			}
			fmt.Println("Sent PSYNC to master")

			// Read FULLRESYNC response
			_, err = conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading FULLRESYNC:", err)
				return
			}

			// Read RDB file (skip the RDB data for now)
			rdbBuffer := make([]byte, 4096)
			_, err = conn.Read(rdbBuffer)
			if err != nil {
				fmt.Println("Error reading RDB:", err)
				return
			}

		}()
	}
}
