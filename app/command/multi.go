package command

func handleMULTI(conn net.Conn, args []string) {
	isMulti[conn] = true
	utils.WriteSimpleString(conn, "OK")
}
