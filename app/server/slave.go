package server

type Slave struct {
	*Server
	Master Master
}
