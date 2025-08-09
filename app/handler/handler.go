package handler

import (
	"SkylerRedis/app/command"
	"SkylerRedis/app/utils"
	"SkylerRedis/app/server"
	"bufio"
	"net"
)

func HandleConnection(conn net.Conn, server server.Server) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		args, err := utils.ParseArgs(conn, reader)
		if err != nil {
			utils.WriteError(conn, err.Error())
			continue
		}
		if len(args) == 0 {
			utils.WriteError(conn, "empty command")
			return
		}
		command.HandleCommand(conn, args, server)
	}
}

// Helpers
