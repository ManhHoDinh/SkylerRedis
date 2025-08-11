package server

type Master struct {
	*Server
	Slaves []Slave
}
