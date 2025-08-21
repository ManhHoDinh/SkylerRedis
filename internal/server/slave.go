package server

import (
	"SkylerRedis/internal/command"
	"SkylerRedis/internal/entity"
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Slave struct {
	*entity.BaseServer
}

func (Slave) HandleConnection(Conn net.Conn) {
	reader := bufio.NewReader(Conn)
	for {
		args, err := utils.ParseArgs(Conn, reader)
		if err != nil {
			return
		}
		if len(args) == 0 {
			utils.WriteError(Conn, "empty command")
			return
		}
		go command.HandleCommand(Conn, args, false)
	}
}
func (Slave) HandShake(replicaof *string, port *string) {
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
		err = handCommand(conn, []string{"PING"}, readerFromMaster, "PONG")
		if err != nil {
			return
		}
		// Send first REPLCONF
		err = handCommand(conn, []string{"REPLCONF", "listening-port", *port}, readerFromMaster, "first REPLCONF")
		if err != nil {
			return
		}
		// Send second REPLCONF
		err = handCommand(conn, []string{"REPLCONF", "ncapa", "psync2"}, readerFromMaster, "Second REPLCONF")
		if err != nil {
			return
		}
		// Send PSYNC
		err = handCommand(conn, []string{"PSYNC", "$1", "$2", "-1"}, readerFromMaster, "FULLRESYNC")
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
	_, err = readerFromMaster.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading second response message: %s. Error: %v\n", responseMessage, err)
		return err
	}
	fmt.Printf("Sent %s to master\n", args[0])
	return nil
}

func readRDBfromMaster(conn net.Conn, readerFromMaster *bufio.Reader) {
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
		args, byteCount, err := utils.ParseArgsWithByteCount(conn, readerFromMaster)
		if err != nil {
			fmt.Println("Error parsing command from master:", err)
			// Don't return immediately, the connection might recover
			continue
		}

		if len(args) == 0 {
			continue
		}

		fmt.Printf("Received propagated command from master: %v (bytes: %d)\n", args, byteCount)

		// Handle REPLCONF GETACK specially - it needs to respond to the master
		if len(args) >= 2 && strings.ToUpper(args[0]) == "REPLCONF" && strings.ToUpper(args[1]) == "GETACK" {
			// Don't update offset for REPLCONF GETACK - it should return the offset before this command
			// Use the real master connection for REPLCONF GETACK
			command.HandleCommand(conn, args, false)
			// Update offset AFTER handling REPLCONF GETACK
			memory.OffSet += byteCount
			fmt.Printf("Updated offset to %d after REPLCONF GETACK\n", memory.OffSet)
		} else {
			// Update offset BEFORE processing other commands
			memory.OffSet += byteCount
			fmt.Printf("Updated offset to %d after %s command\n", memory.OffSet, args[0])

			// Process the command (it will be applied to local state but no response sent)
			command.HandleCommand(&entity.MockConn{}, args, false)
		}
	}
}
