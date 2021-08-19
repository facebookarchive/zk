package zk

import (
	"context"
	"fmt"
	"net"
	"time"
)

const defaultTimeout = 2 * time.Second

// Client represents a Zookeeper client abstraction with additional configuration parameters.
type Client struct {
	// Dialer is a function to be used to establish a connection to a single host.
	Dialer         func(ctx context.Context, network, addr string) (net.Conn, error)
	SessionTimeout time.Duration

	MaxRetries int
	Network    string
	Ensemble   string

	conn *Conn
}

// GetData uses the retryable client to call Get on a Zookeeper server.
func (client *Client) GetData(ctx context.Context, path string) ([]byte, error) {
	var err error
	var data []byte
	err = client.doRetry(ctx, func() error {
		data, err = client.conn.GetData(path)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("connection failed after %d retries: %w", client.MaxRetries, err)
	}

	return data, nil
}

// GetChildren uses the retryable client to call GetChildren on a Zookeeper server.
func (client *Client) GetChildren(ctx context.Context, path string) ([]string, error) {
	var children []string
	var err error
	err = client.doRetry(ctx, func() error {
		children, err = client.conn.GetChildren(path)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("connection failed after %d retries: %w", client.MaxRetries, err)
	}

	return children, nil
}

// Reset closes the client's underlying connection, cancelling any RPCs currently in-flight.
// Future RPC calls will need to re-initialize the connection.
func (client *Client) Reset() error {
	return client.conn.Close()
}

// doRetry makes attempts at connection and RPC execution according to the MaxRetries parameter.
// If the MaxRetries value is not set, the RPC is executed only once.
func (client *Client) doRetry(ctx context.Context, fun func() error) error {
	var err error
	for i := 0; i <= client.MaxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err() // ctx canceled, don't retry
		}
		if err = client.getConn(ctx); err != nil {
			continue
		}

		err = fun()
		if err != nil {
			continue // retry
		}

		return nil
	}

	return err
}

// getConn initializes client connection or reuses it if it has already been established.
func (client *Client) getConn(ctx context.Context) error {
	if client.conn == nil || !client.conn.isAlive() {
		conn, err := client.DialContext(ctx, client.Network, client.Ensemble)
		if err != nil {
			return err
		}

		client.conn = conn
	}

	return nil
}
