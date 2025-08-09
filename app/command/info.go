package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
	"net"
	"strings"
)

func handleINFO(conn net.Conn, args []string, server server.Server) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'INFO'")
		return
	}
	switch strings.ToUpper(args[1]) {
	case "REPLICATION":
		if server.IsMaster {
			utils.WriteBulkString(conn, "role:master")
		} else {
			utils.WriteBulkString(conn, "role:slave")
		}
	default:
		utils.WriteError(conn, fmt.Sprintf("unknown INFO section '%s'", args[1]))
	}
}
