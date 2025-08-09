package main

import (
	"SkylerRedis/app/handler"
	"SkylerRedis/app/server"
	"flag"
	"fmt"
	"net"
	"os"
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
	} else{
		Server.IsMaster = true
	}
	fmt.Println("Server is running on port", *port)
	fmt.Println("Replica of:", *replicaof)
	if !Server.IsMaster {
		fmt.Println("Running as a replica of", *replicaof)
	} else {
		fmt.Println("Running as master")
	}


	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handler.HandleConnection(conn, Server)
	}
}
