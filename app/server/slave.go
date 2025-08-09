package server


type Slave struct {
	*Server
	MasterAddr string
	IsConnected bool
}
