package server

import (
	"SkylerRedis/internal/command"
	"SkylerRedis/internal/entity"
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type Slave struct {
	*entity.BaseServer
	ReplicaOf *string
	Port      *string
}

// IsMaster returns false for the Slave server type.
func (s *Slave) IsMaster() bool {
	return false
}

// HandShake connects to the master and initiates the replication process.
func (s *Slave) HandShake() {
	if s.ReplicaOf == nil || *s.ReplicaOf == "" {
		fmt.Println("Master address not provided for slave.")
		return
	}

	parts := strings.Split(*s.ReplicaOf, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid replicaof format. Expected 'host port'.")
		return
	}

	masterAddr := parts[0] + ":" + parts[1]
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Println("Failed to connect to master:", err)
		return
	}

	fmt.Println("Connected to master:", masterAddr)

	// This goroutine handles all communication with the master.
	// It's separate from the main client-facing event loop.
	go func() {
		defer conn.Close()
		readerFromMaster := bufio.NewReader(conn)

		// Send PING
		err = handCommand(conn, []string{"PING"}, readerFromMaster, "PONG")
		if err != nil {
			return
		}
		// Send first REPLCONF
		err = handCommand(conn, []string{"REPLCONF", "listening-port", *s.Port}, readerFromMaster, "first REPLCONF")
		if err != nil {
			return
		}
		// Send second REPLCONF
		err = handCommand(conn, []string{"REPLCONF", "capa", "psync2"}, readerFromMaster, "Second REPLCONF")
		if err != nil {
			return
		}
		// Send PSYNC
		err = handCommand(conn, []string{"PSYNC", "?", "-1"}, readerFromMaster, "PSYNC")
		if err != nil {
			return
		}
		readRDBfromMaster(conn, readerFromMaster)
	}()
}

func handCommand(conn net.Conn, args []string, readerFromMaster *bufio.Reader, responseMessage string) error {
	_, err := conn.Write([]byte(utils.ArgsToRESP(args)))
	if err != nil {
		fmt.Println("Connection to master lost:", err)
		return err
	}
	// It's better to read the response to confirm.
	// For simplicity, we assume it succeeds if write succeeds.
	// In a real scenario, you'd parse the response from readerFromMaster.
	fmt.Printf("Sent %s to master\n", args[0])
	return nil
}

func readRDBfromMaster(conn net.Conn, readerFromMaster *bufio.Reader) {
	// Read RDB file - we need to parse the length and then read that many bytes
	line, err := readerFromMaster.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading RDB length line:", err)
		return
	}

	if line[0] == '+' { // Check for "+FULLRESYNC..."
		fmt.Println("Full resync with master started.")
		// The next line should be the RDB length
		line, err = readerFromMaster.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading RDB length line after FULLRESYNC:", err)
			return
		}
	}


	var rdbLength int
	if _, err := fmt.Sscanf(line, "$%d\r\n", &rdbLength); err != nil {
		fmt.Printf("Error parsing RDB length from %q: %v\n", line, err)
		return
	}

	// Read exactly that many bytes for RDB data
	rdbData := make([]byte, rdbLength)
	if _, err := io.ReadFull(readerFromMaster, rdbData); err != nil {
		fmt.Println("Error reading RDB data:", err)
		return
	}


	fmt.Printf("RDB file received (%d bytes), now listening for propagated commands...\n", rdbLength)

	// Continue listening for propagated commands from master
	for {
		args, byteCount, err := utils.ParseArgsWithByteCount(conn, readerFromMaster)
		if err != nil {
			fmt.Println("Error parsing command from master:", err)
			return // Connection is likely lost
		}

		if len(args) == 0 {
			continue
		}
		
		isGetAck := len(args) >= 2 && strings.ToUpper(args[0]) == "REPLCONF" && strings.ToUpper(args[1]) == "GETACK"

		if isGetAck {
			// Respond to master with current offset
			// For REPLCONF GETACK, there's no specific key for sharding data.
			// It's a control command, so we can route it to shard 0.
			command.HandleCommand(conn, args, false, memory.Shards[0]) 
		} else {
			// Apply propagated command to local state
			// Get shard based on the command's key
			var key string
			if len(args) > 1 {
				key = args[1]
			}
			shard := memory.GetShardForKey(key)
			command.HandleCommand(&entity.MockConn{}, args, false, shard) 
		}
		
		// Always update the offset
		memory.OffSet += byteCount
	}
}
