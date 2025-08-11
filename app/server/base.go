package server

type Server struct {
	SeverId     int
	Addr        string
	IsMaster    bool
	IsConnected bool
}
