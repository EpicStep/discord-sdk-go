package transport

import (
	"fmt"
	"net"
)

// Listener ...
type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

type listener struct {
	listener net.Listener
}

func (l *listener) Accept() (Conn, error) {
	c, err := l.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("l.Accept: %w", err)
	}

	return newConn(c), nil
}

func (l *listener) Close() error {
	return l.listener.Close()
}

func (l *listener) Addr() net.Addr {
	return l.listener.Addr()
}

// Listen on IPC.
func Listen(instanceID uint) (Listener, error) {
	l, err := listenIPC(getDiscordFilename(instanceID))
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	return &listener{
		listener: l,
	}, nil
}
