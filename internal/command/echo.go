package command

import (
	"SkylerRedis/internal/utils"
	"net"
)

type Echo struct{}

func (Echo) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'ECHO'")
		return
	}
	utils.WriteSimpleString(Conn, args[1])
}
