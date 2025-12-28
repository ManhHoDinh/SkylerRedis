package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type MULTI struct{}

func (MULTI) Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	memory.IsMulti[Conn] = true
	utils.WriteSimpleString(Conn, "OK")
}
