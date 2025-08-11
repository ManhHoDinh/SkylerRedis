package handler

import (
	"SkylerRedis/app/command"
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"bufio"
	"fmt"
)

func HandleConnection(server server.Server) {
	defer server.Conn.Close()
	reader := bufio.NewReader(server.Conn)
	for {
		args, err := utils.ParseArgs(server.Conn, reader)
		if err != nil {
			utils.WriteError(server.Conn, err.Error())
			continue
		}
		if len(args) == 0 {
			utils.WriteError(server.Conn, "empty command")
			return
		}
		command.HandleCommand(server, args)
		if server.IsMaster {
			fmt.Println("Master detected, forwarding command to slaves")
			fmt.Println("Master:", memory.Master)
			fmt.Println("Slaves:", memory.Master.Slaves)
			for _, ser := range memory.Master.Slaves {
				fmt.Println("Forwarding command to slave:", ser.Addr)
				command.HandleCommand(*ser.Server, args)
			}
		}
	}
}

// Helpers
