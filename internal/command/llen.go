package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type LLen struct{}

func (LLen) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'LLEN'")
		return
	}
	utils.WriteInteger(Conn, len(memory.RPush[args[1]]))
}
