package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"net"
)

func handlePSYNC(conn net.Conn, args []string, server server.Server) {
	if !server.IsMaster {
		utils.WriteError(conn, "PSYNC can only be called by master")
		return
	}
}
