package entity

import (
	"net"
	"time"
)

type MockConn struct{}

func (m *MockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *MockConn) Write(b []byte) (n int, err error)  { return len(b), nil } // Pretend to write successfully
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }
