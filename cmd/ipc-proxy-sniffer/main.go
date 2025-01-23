// Package main contains an IPC proxy used for sniffing the communication between the original C++ library and the Discord client network.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/EpicStep/discord-sdk-go/transport"
)

var (
	listenInstanceID = flag.Uint("listen-instance-id", 0, "instance id for listening")
	originInstanceID = flag.Uint("origin-instance-id", 1, "origin discord instance id")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Println(err)
	}
}

func run(ctx context.Context) error {
	listener, err := transport.Listen(*listenInstanceID)
	if err != nil {
		return err
	}

	fmt.Println("Listening on", listener.Addr())

	defer func() {
		if err = listener.Close(); err != nil {
			log.Println("Failed to close listener:", err)
			return
		}
	}()

	fmt.Println("Waiting client connection...")

	listenerErrCh := make(chan error, 1)
	listenerDoneCh := make(chan struct{})

	connectionsWG := new(sync.WaitGroup)

	go func() {
		defer close(listenerDoneCh)

		var currentConnectionID uint

		for {
			conn, listenerErr := listener.Accept()
			if listenerErr != nil {
				listenerErrCh <- listenerErr
				return
			}

			currentConnectionID++

			go handleConn(ctx, connectionsWG, currentConnectionID, conn)
		}
	}()

	select {
	case err = <-listenerErrCh:
		return fmt.Errorf("listener error: %w", err)
	case <-ctx.Done():
	}

	fmt.Println("Closing listener...")

	if err = listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:contextcheck
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("listener close timeout: %w", ctx.Err())
	case <-listenerDoneCh:
	}

	fmt.Println("Listener is closed")

	return nil
}

func handleConn(ctx context.Context, wg *sync.WaitGroup, connectionID uint, clientConn transport.Conn) {
	wg.Add(1)
	defer wg.Done()

	fmt.Printf("Client is connected, id %d", connectionID)
	defer fmt.Printf("Client is disconnected, id %d", connectionID)

	originConn, err := transport.Dial(ctx, transport.DialOptions{
		InstanceID: *originInstanceID,
	})
	if err != nil {
		fmt.Printf("[%d] Failed to open connection to origin: %s\n", connectionID, err)
		return
	}

	eg, eCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return writer(eCtx, connectionID, clientConn, originConn)
	})

	eg.Go(func() error {
		return reader(eCtx, connectionID, clientConn, originConn)
	})

	eg.Go(func() error {
		<-eCtx.Done()
		return errors.Join(originConn.Close(), clientConn.Close())
	})

	if err = eg.Wait(); err != nil {
		fmt.Printf("[%d] eg.Wait: %s\n", connectionID, err)
		return
	}
}

func reader(ctx context.Context, connectionID uint, client, origin transport.Conn) error {
	for {
		opcode, data, err := client.Read(ctx)
		if err != nil {
			err = checkCtxError(ctx, err)
			if err == nil {
				continue
			}

			return fmt.Errorf("failed to read data from client: %w", err)
		}

		fmt.Printf("[%d] [send] opcode: %d, data: %s\n", connectionID, opcode, string(data))

		if err = origin.Write(ctx, opcode, data); err != nil {
			return fmt.Errorf("failed to write data to origin: %w", err)
		}
	}
}

func writer(ctx context.Context, connectionID uint, client, origin transport.Conn) error {
	for {
		opcode, data, err := origin.Read(ctx)
		if err != nil {
			err = checkCtxError(ctx, err)
			if err == nil {
				continue
			}

			return fmt.Errorf("failed to read data from origin: %w", err)
		}

		fmt.Printf("[%d] [recieve] opcode: %d, data: %s\n", connectionID, opcode, string(data))

		if err = client.Write(ctx, opcode, data); err != nil {
			return fmt.Errorf("failed to write data to client: %w", err)
		}
	}
}

func checkCtxError(ctx context.Context, err error) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("writer: %w", ctx.Err())
	default:
		if noUpdates(err) {
			return nil
		}
	}

	return err
}

func noUpdates(err error) bool {
	var e *net.OpError
	if errors.As(err, &e) && e.Timeout() {
		return true
	}
	return false
}
