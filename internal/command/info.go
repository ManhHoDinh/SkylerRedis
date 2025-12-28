package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"strings"
)

type INFO struct{}

// HandleHandle implements ICommand.
func (i INFO) HandleHandle(Conn net.Conn, args []string, isMaster bool) {
	panic("unimplemented")
}

// HandleHandle implements ICommand.
func (INFO) Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'INFO'")
		return
	}
	switch strings.ToUpper(args[1]) {
	case "REPLICATION":
		if isMaster {
			utils.WriteBulkString(Conn,
				fmt.Sprintf("role:master\nmaster_replid:%s\nmaster_repl_offset:0", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"))
		} else {
			utils.WriteBulkString(Conn, "role:slave")
		}
	default:
		utils.WriteError(Conn, fmt.Sprintf("unknown INFO section '%s'", args[1]))
	}
}
