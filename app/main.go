package main

import (
	"SkylerRedis/app/handler"
	"SkylerRedis/app/memory"
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

func main() {
	flag.Parse() // Parse command line arguments
	l, err := net.Listen("tcp", "0.0.0.0:"+*port)
	if err != nil {
		fmt.Println("Failed to bind to port", *port)
		os.Exit(1)
	}
	Server := server.Server{}
	if *replicaof != "" {
		fmt.Println("Server is running on port", *port)
		fmt.Println("Replica of:", *replicaof)

		Server.SeverId = len(memory.Master.Slaves) + 1
		Server.Addr = "slave:" + *port
		slave := server.Slave{
			Master: memory.Master,
			Server: &Server,
		}
		if len(memory.Master.Slaves) == 0 {
			memory.Master.Slaves = make([]*server.Slave, 0)
		}
		memory.Master.Slaves = append(memory.Master.Slaves, &slave)
		fmt.Println("Slave ID:", Server.SeverId)
		fmt.Println("Slaves:", memory.Master.Slaves)
	} else {
		fmt.Println("Running as master")
		Server.SeverId = 0
		Server.IsMaster = true
		Server.Addr = "master:" + *port
		memory.Master.Server = &Server
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Failed to accept connection:", err)
				continue
			}
			Server.Conn = conn
			fmt.Println("Master:", memory.Master)
			go handler.HandleConnection(Server)
		}
	}()
	go sendToMaster()
	select {} // Keep the main goroutine running
}
func sendToMaster() {
	if *replicaof != "" {
		parts := strings.Split(*replicaof, " ")
		masterAddr := parts[0] + ":" + parts[1]
		conn, err := net.Dial("tcp", masterAddr)
		if err != nil {
			fmt.Println("Failed to connect to master:", err)
			time.Sleep(5 * time.Second)
		}

		fmt.Println("Connected to master:", masterAddr)
		_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
		if err != nil {
			fmt.Println("Connection to master lost:", err)
			conn.Close()
		}
		fmt.Println("Sent PING to master")
		time.Sleep(5 * time.Millisecond)
		_, err = conn.Write([]byte(fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%s\r\n", *port)))

		if err != nil {
			fmt.Println("Connection to master lost:", err)
			conn.Close()
		}
		fmt.Println("Sent first REPLCONF to master")
		time.Sleep(5 * time.Millisecond)
		_, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
		if err != nil {
			fmt.Println("Connection to master lost:", err)
			conn.Close()
		}
		fmt.Println("Sent Second REPLCONF to master")

		time.Sleep(5 * time.Millisecond)
		_, err = conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
		if err != nil {
			fmt.Println("Connection to master lost:", err)
			conn.Close()
		}
		fmt.Println("Sent PSYNC to master")
		time.Sleep(5 * time.Millisecond)
		defer conn.Close()
	}
}
