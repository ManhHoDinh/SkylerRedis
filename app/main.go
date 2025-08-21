package main

import (
	"SkylerRedis/internal/server"
	"flag"
	"fmt"
	"net"
	"os"
)

var port = flag.String("port", "6379", "Port for redis server")
var replicaof = flag.String("replicaof", "", "Address of the master server")
var hostname = flag.String("hostname", "0.0.0.0", "Hostname for the server")

func main() {
	flag.Parse() // Parse command line arguments
	Server := intServer(replicaof, port)
	go handleConnection(Server)
	select {} // Keep the main goroutine running
}
func intServer(replicaof *string, port *string) server.Server {
	var Server server.Server
	if *replicaof != "" {
		fmt.Println("Server is running on port", *port)
		fmt.Println("Replica of:", *replicaof)
		Server = server.Slave{}
		go Server.HandShake(replicaof, port)
	} else {
		fmt.Println("Running as master")
		Server = server.Master{}
	}
	return Server
}

func handleConnection(Server server.Server) {
	l, err := net.Listen("tcp", *hostname+":"+*port)
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
		go Server.HandleConnection(conn)
	}
}
