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
			utils.WriteBulkString(conn,
				fmt.Sprintf("role:master\nmaster_replid:%s\nmaster_repl_offset:0", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"))
		} else {
			utils.WriteBulkString(conn, "role:slave")
		}
	default:
		utils.WriteError(conn, fmt.Sprintf("unknown INFO section '%s'", args[1]))
	}
}
