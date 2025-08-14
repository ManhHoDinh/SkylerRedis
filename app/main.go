package main

import (
	"SkylerRedis/app/command"
	"SkylerRedis/app/handler"
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
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
			// Use a single buffered reader for all communication
			readerFromMaster := bufio.NewReader(conn)
			
			// Send PING
			_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
			if err != nil {
				fmt.Println("Connection to master lost:", err)
				return
			}
			fmt.Println("Sent PING to master")

			// Read PONG response
			_, err = readerFromMaster.ReadString('\n')
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
			_, err = readerFromMaster.ReadString('\n')
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
			_, err = readerFromMaster.ReadString('\n')
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
			_, err = readerFromMaster.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading FULLRESYNC:", err)
				return
			}

			// Read RDB file - we need to parse the length and then read that many bytes
			lengthLine, err := readerFromMaster.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading RDB length:", err)
				return
			}
			
			// Parse RDB length from $88\r\n format
			var rdbLength int
			if _, err := fmt.Sscanf(lengthLine, "$%d", &rdbLength); err != nil {
				fmt.Printf("Error parsing RDB length from %q: %v\n", lengthLine, err)
				return
			}
			
			// Read exactly that many bytes for RDB data
			rdbData := make([]byte, rdbLength)
			_, err = readerFromMaster.Read(rdbData)
			if err != nil {
				fmt.Println("Error reading RDB data:", err)
				return
			}

			fmt.Printf("RDB file received (%d bytes), now listening for propagated commands...\n", rdbLength)

			// Continue listening for propagated commands from master using the same reader
			for {
				args, err := utils.ParseArgs(conn, readerFromMaster)
				if err != nil {
					fmt.Println("Error parsing command from master:", err)
					// Don't return immediately, the connection might recover
					continue
				}
				
				if len(args) == 0 {
					continue
				}

				fmt.Printf("Received propagated command from master: %v\n", args)

				// Create a mock server for processing the command without sending responses
				mockServer := server.Server{
					Conn:     &mockConn{}, // Use a mock connection that doesn't actually write
					IsMaster: false,
				}
				
				// Process the command (it will be applied to local state but no response sent)
				command.HandleCommand(mockServer, args)
				
				// Debug: check if the value was stored
				if len(args) >= 3 && strings.ToUpper(args[0]) == "SET" {
					fmt.Printf("After SET %s=%s, checking storage...\n", args[1], args[2])
					if entry, exists := memory.Store[args[1]]; exists {
						fmt.Printf("Key %s stored with value: %q\n", args[1], entry.Value)
					} else {
						fmt.Printf("Key %s NOT found in storage!\n", args[1])
					}
				}
			}

		}()
	}
}

// mockConn implements net.Conn interface but doesn't actually send data
// This is used for processing propagated commands without sending responses back to master
type mockConn struct{}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error) { return len(b), nil } // Pretend to write successfully
func (m *mockConn) Close() error                      { return nil }
func (m *mockConn) LocalAddr() net.Addr               { return nil }
func (m *mockConn) RemoteAddr() net.Addr              { return nil }
func (m *mockConn) SetDeadline(t time.Time) error     { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
