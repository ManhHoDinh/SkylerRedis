package command

func handlePing(conn net.Conn) {
	utils.WriteSimpleString(conn, "PONG")
}
