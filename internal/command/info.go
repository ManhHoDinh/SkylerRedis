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
func (INFO) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'INFO'")
		return
	}
	switch strings.ToUpper(args[1]) {
	case "REPLICATION":
		if isMaster {
			utils.WriteBulkString(Conn,
				fmt.Sprintf("role:master\nmaster_replid:%s\nmaster_repl_offset:%d\nconnected_slaves:%d",
					masterReplID, masterReplOffset, connectedSlaves))
		} else {
			utils.WriteBulkString(Conn, "role:slave")
		}
	default:
		utils.WriteError(Conn, fmt.Sprintf("unknown INFO section '%s'", args[1]))
	}
}
