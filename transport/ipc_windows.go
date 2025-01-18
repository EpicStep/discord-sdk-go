//go:build windows

package transport

import (
	"context"
	"fmt"
	"net"

	"github.com/Microsoft/go-winio"
)

func listenIPC(filename string) (net.Listener, error) {
	l, err := winio.ListenPipe(`\\.\pipe\`+filename, &winio.PipeConfig{
		InputBufferSize:  maxMessageSize,
		OutputBufferSize: maxMessageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", filename, err)
	}

	return l, nil
}

func openConn(ctx context.Context, _ net.Dialer, filename string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, `\\.\pipe\`+filename)
}
