package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Type struct{}

func (Type) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'TYPE'")
		return
	}
	key := args[1]
	if _, exists := memory.RPush[key]; exists {
		utils.WriteSimpleString(Conn, "list")
	} else if _, exists := memory.Store[key]; exists {
		utils.WriteSimpleString(Conn, "string")
	} else if _, exists := memory.Stream[key]; exists {
		utils.WriteSimpleString(Conn, "stream")
	} else {
		utils.WriteSimpleString(Conn, "none")
	}
}
