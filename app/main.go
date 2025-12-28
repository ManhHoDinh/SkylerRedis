package main

import (
	"SkylerRedis/internal/command"
	"SkylerRedis/internal/eventloop"
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/server"
	"SkylerRedis/internal/utils"
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings" // Added for strings.ToUpper
	"time"
)

var port = flag.String("port", "6379", "Port for redis server")
var replicaof = flag.String("replicaof", "", "Address of the master server")
var hostname = flag.String("hostname", "0.0.0.0", "Hostname for the server")
var maxmemory = flag.Int("maxmemory", 0, "Max memory in number of keys (0 for no limit)")
var numShards = flag.Int("numshards", 1, "Number of data shards") // New flag for sharding

func main() {
	flag.Parse() // Parse command line arguments

	var replicaOfStr string
	if *replicaof != "" {
		// CodeCrafters runner passes --replicaof host port as separate arguments.
		// The flag package will assign 'host' to *replicaof, and 'port' will be a positional arg.
		if len(flag.Args()) > 0 {
			replicaOfStr = *replicaof + " " + flag.Args()[0]
		} else {
			// Handle cases where only --replicaof host is provided, though this is unlikely for CodeCrafters
			log.Println("Warning: --replicaof flag used without a port argument.")
			replicaOfStr = *replicaof
		}
	}
	
	// Initialize shards before anything else.
	memory.InitShards(*numShards, *maxmemory)

	serverInstance := intServer(&replicaOfStr, port)

	// The ReadCallback is the heart of handling client data.
	// It's called by the event loop when a connection has data to be read.
	readCallback := func(conn net.Conn) error {
		reader := bufio.NewReader(conn)
		args, err := utils.ParseArgs(conn, reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected.", conn.RemoteAddr().String())
				return err // Signal to the event loop to close the connection.
			}
			log.Printf("Error parsing command from %s: %v", conn.RemoteAddr().String(), err)
			return err
		}

		// Handle REPLCONF LISTENING-PORT directly to break import cycle
		if len(args) >= 3 && strings.ToUpper(args[0]) == "REPLCONF" && strings.ToUpper(args[1]) == "LISTENING-PORT" {
			if serverInstance.IsMaster() {
				err := serverInstance.RegisterSlave(conn, args[2])
				if err != nil {
					utils.WriteError(conn, fmt.Sprintf("ERR failed to register slave: %v", err))
				} else {
					utils.WriteSimpleString(conn, "OK")
				}
			} else {
				utils.WriteError(conn, "ERR REPLCONF command only accepted by master server")
			}
			return nil
		}

		// Determine the shard based on the key (first argument after command for most cases)
		// For now, it always returns shard 0.
		var key string
		if len(args) > 1 {
			key = args[1]
		}
		shard := memory.GetShardForKey(key) // Get the appropriate shard for this command

		isMaster := serverInstance.IsMaster()
		var masterReplID string
		var masterReplOffset int
		var connectedSlaves int
		if isMaster {
			master, ok := serverInstance.(*server.Master)
			if ok { // Should always be true if serverInstance.IsMaster() is true
				masterReplID = master.MasterReplID
				masterReplOffset = master.MasterReplOffset
				connectedSlaves = len(master.Slaves)
			}
		}
		
		command.HandleCommand(conn, args, isMaster, masterReplID, masterReplOffset, connectedSlaves, shard)

		if serverInstance.IsMaster() && utils.IsModifyCommand(args) {
			master, ok := serverInstance.(*server.Master)
			if ok { // Should always be true if serverInstance.IsMaster() is true
				go master.PropagateCommand(args) // Propagate in a goroutine to avoid blocking the event loop
			}
		}

		// After handling a command that might add data, check for LRU eviction on the specific shard
		// The EvictKeysByLRU function now contains its own check for MaxMemory limit
		shard.EvictKeysByLRU() 

		return nil // Success, keep connection open.
	}

	el, err := eventloop.New(readCallback)
	if err != nil {
		log.Fatalf("Failed to create event loop: %v", err)
	}

	// Start the event loop in a background goroutine.
	go el.Start()

	// Start the background task for expiring keys and LRU eviction for each shard.
	// Each shard now manages its own eviction.
	memory.ShardsMu.RLock() // Protect reading from memory.Shards
	for i := 0; i < *numShards; i++ {
		shard := memory.Shards[i]
		go func(s *memory.Shard) {
			ticker := time.NewTicker(100 * time.Millisecond)
			for range ticker.C {
				s.EvictRandomKeys() // Evict expired keys
				s.EvictKeysByLRU()  // Perform LRU eviction if memory limit is active
			}
		}(shard)
	}
	memory.ShardsMu.RUnlock() // Release read lock

	// Create the TCP listener.
	l, err := net.Listen("tcp", *hostname+":"+*port)
	if err != nil {
		log.Fatalf("Failed to bind to port %s: %v", *port, err)
	}
	log.Printf("SkylerRedis is listening on port %s", *port)

	// Main loop to accept new connections.
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted connection from %s", conn.RemoteAddr().String())

		// Add the new connection to the event loop for I/O polling.
		if err := el.Add(conn); err != nil {
			log.Printf("Failed to add connection to event loop: %v", err)
			conn.Close()
		}
	}
}

// intServer now returns a more concrete type to access IsMaster, etc.
func intServer(replicaof *string, port *string) server.Server {
	if *replicaof != "" {
		fmt.Println("Server is running as SLAVE on port", *port)
		fmt.Println("Replica of:", *replicaof)
		slave := server.NewSlave(replicaof, port)
		go slave.HandShake()
		return slave
	} else {
		fmt.Println("Running as MASTER")
		return server.NewMaster()
	}
}
