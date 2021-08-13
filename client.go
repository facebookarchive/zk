package zk

import (
	"context"
	"fmt"
	"net"
	"time"
)

const defaultMaxRetries = 5
const defaultTimeout = 2 * time.Second

// Client represents a Zookeeper client abstraction with additional configuration parameters.
type Client struct {
	// Dialer is a function to be used to establish a connection to a single host.
	Dialer         func(ctx context.Context, network, addr string) (net.Conn, error)
	SessionTimeout time.Duration

	Network           string
	EnsembleAddresses []string
	MaxRetries        int

	conn zkConn
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

func (client *Client) doRetry(ctx context.Context, fun func() error) error {
	if client.MaxRetries == 0 {
		client.MaxRetries = defaultMaxRetries
	}

	var err error
	for i := 0; i < client.MaxRetries; i++ {
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

// tryDial attempts to dial all of the servers in a Client's ensemble until a successful connection is established.
func (client *Client) tryDial(ctx context.Context) (zkConn, error) {
	var conn *Conn
	var err error
	shuffleSlice(client.EnsembleAddresses)
	for _, address := range client.EnsembleAddresses {
		if ctx.Err() != nil {
			return nil, ctx.Err() // canceled, don't retry
		}

		conn, err = client.DialContext(ctx, client.Network, address)
		if err != nil {
			continue // try dialing the next address in the ensemble
		}

		return conn, nil
	}

	return nil, fmt.Errorf("could not dial any servers in ensemble: %w", err)
}

// getConn initializes client connection or reuses it if it has already been established.
func (client *Client) getConn(ctx context.Context) error {
	if client.conn == nil || !client.conn.isAlive() {
		conn, err := client.tryDial(ctx)
		if err != nil {
			return err
		}

		client.conn = conn
	}

	return nil
}
