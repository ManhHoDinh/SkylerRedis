package command

import (
	"SkylerRedis/app/utils"
	"net"
	"time"
)

func handlePSYNC(conn net.Conn, args []string) {
	utils.WriteSimpleString(conn, "FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0")
	time.Sleep(5 * time.Millisecond)
	utils.WriteBulkString(conn, "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog==")
	time.Sleep(5 * time.Millisecond)
}
