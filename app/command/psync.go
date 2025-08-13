package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"bufio"
	"fmt"
	"io"
)

func handlePSYNC(server server.Server, args []string) {
	// 1. Acknowledge FULLRESYNC
	utils.WriteSimpleString(server.Conn, "FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0")

	// 2. Read the RDB bulk string header: $<len>
	lenLine, _ := bufio.NewReader(server.Conn).ReadString('\n')
	var rdbLen int
	fmt.Sscanf(lenLine, "$%d", &rdbLen)

	// 3. Read and discard the RDB content
	buf := make([]byte, rdbLen)
	io.ReadFull(server.Conn, buf)

	// 4. Now enter loop to read propagated commands from master
	for {
		args, err := utils.ParseArgs(server.Conn, bufio.NewReader(server.Conn))
		if err != nil {
			fmt.Println("Error reading propagated command:", err)
			return
		}
		if len(args) == 0 {
			continue
		}
		// Apply to our DB, but DO NOT send any reply to master
		HandleCommand(server, args)
	}
}
