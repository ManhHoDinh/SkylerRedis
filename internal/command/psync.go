package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"encoding/hex"
	"fmt"
	"net"
)

type PSYNC struct{}

func (PSYNC) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if !isMaster {
		utils.WriteError(Conn, "ERR PSYNC command only accepted by master server")
		return
	}

	utils.WriteSimpleString(Conn, fmt.Sprintf("FULLRESYNC %s %d", masterReplID, masterReplOffset))
	// Hardcoded RDB content for an empty RDB file, as typically expected by CodeCrafters
	RDBcontent, _ := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
	Conn.Write([]byte(fmt.Sprintf("$%v\r\n%v", len(string(RDBcontent)), string(RDBcontent))))
}
