package transport

import (
	"context"
	"net"
)

// DialOptions ...
type DialOptions struct {
	// Dialer used to connect to IPC (only on unix).
	Dialer net.Dialer
	// EnableInstanceLookup enables automatic Discord client lookup.
	EnableInstanceLookup bool
	// InstanceID is variable that you can use to handle specific Discord clients. Used only if EnableInstanceLookup is false.
	// Alternative to DISCORD_INSTANCE_ID.
	// https://discord.com/developers/docs/game-sdk/getting-started#testing-locally-with-two-clients-environment-variable-example
	InstanceID uint
}

// Dial connects to IPC and returns Conn.
func Dial(ctx context.Context, opts DialOptions) (Conn, error) {
	openSpecificConn := func(instanceID uint) (net.Conn, error) {
		return openConn(ctx, opts.Dialer, getDiscordFilename(instanceID))
	}

	var c net.Conn
	var err error
	if opts.EnableInstanceLookup {
		for i := uint(0); i < 3; i++ {
			c, err = openSpecificConn(i)
			if err == nil {
				break
			}
		}
	} else {
		c, err = openConn(ctx, opts.Dialer, getDiscordFilename(opts.InstanceID))
	}

	if err != nil {
		return nil, err
	}

	return newConn(c), nil
}
