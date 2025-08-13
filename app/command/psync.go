package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"encoding/hex"
	"fmt"
)

func handlePSYNC(server server.Server, args []string) {
	utils.WriteSimpleString(server.Conn, "FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0")
	RDBcontent, _ := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
	server.Conn.Write([]byte(fmt.Sprintf("$%v\r\n%v", len(string(RDBcontent)), string(RDBcontent))))
}
