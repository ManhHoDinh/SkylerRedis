//go:build linux

package eventloop

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

// EventLoop represents the core I/O event loop.
// It uses epoll to manage a large number of client connections efficiently.
type EventLoop struct {
	epfd         int
	connections  map[int]net.Conn
	ReadCallback func(conn net.Conn) error
}

// New creates a new EventLoop.
// It initializes the epoll instance.
func New(readCallback func(conn net.Conn) error) (*EventLoop, error) {
	// Create an epoll file descriptor.
	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &EventLoop{
		epfd:         epfd,
		connections:  make(map[int]net.Conn),
		ReadCallback: readCallback,
	}, nil
}

// Start runs the event loop.
func (el *EventLoop) Start() {
	events := make([]unix.EpollEvent, 1024)
	for {
		// Wait for events. -1 means block indefinitely.
		n, err := unix.EpollWait(el.epfd, events, -1)
		if err != nil {
			// If the syscall was interrupted, just continue.
			if err == unix.EINTR {
				continue
			}
			fmt.Printf("EpollWait error: %v\n", err)
			continue
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)
			conn, ok := el.connections[fd]
			if !ok {
				// This might happen if the connection was closed.
				continue
			}

			// Handle read events (EPOLLIN).
			// We also handle EPOLLHUP and EPOLLERR to detect closed connections.
			if events[i].Events&(unix.EPOLLIN|unix.EPOLLHUP|unix.EPOLLERR) != 0 {
				if el.ReadCallback != nil {
					if err := el.ReadCallback(conn); err != nil {
						// If callback returns an error (e.g., EOF), close the connection.
						el.Remove(conn)
					}
				}
			}
		}
	}
}

// Add registers a new connection with the event loop.
func (el *EventLoop) Add(conn net.Conn) error {
	fd, err := connToFileDescriptor(conn)
	if err != nil {
		return err
	}

	// Set the file descriptor to non-blocking mode.
	if err := unix.SetNonblock(fd, true); err != nil {
		return err
	}

	el.connections[fd] = conn
	event := &unix.EpollEvent{
		// Watch for read events and enable edge-triggered mode (EPOLLET).
		Events: unix.EPOLLIN | unix.EPOLLET,
		Fd:     int32(fd),
	}

	return unix.EpollCtl(el.epfd, unix.EPOLL_CTL_ADD, fd, event)
}

// Remove unregisters a connection from the event loop.
func (el *EventLoop) Remove(conn net.Conn) error {
	fd, err := connToFileDescriptor(conn)
	if err != nil {
		// If we can't get the fd, we can't remove it from epoll, but we can close it.
		conn.Close()
		return err
	}

	// Remove from epoll.
	err = unix.EpollCtl(el.epfd, unix.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		fmt.Printf("Failed to remove fd %d from epoll: %v\n", fd, err)
	}
	
	// Close the connection and remove from our map.
	conn.Close()
	delete(el.connections, fd)
	fmt.Printf("Connection from %s closed.\n", conn.RemoteAddr().String())
	return nil
}


// connToFileDescriptor extracts the file descriptor from a net.Conn using syscall.RawConn.
func connToFileDescriptor(conn net.Conn) (int, error) {
	var fileDescriptor int
	var err error

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return 0, fmt.Errorf("connection is not a TCPConn")
	}

	// Get the underlying file descriptor
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return 0, err
	}

	err = rawConn.Control(func(fd uintptr) {
		fileDescriptor = int(fd)
	})
	if err != nil {
		return 0, err
	}

	return fileDescriptor, nil
}
