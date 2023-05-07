// conn/conn.go
// Package conn provides an interface for interacting with TLS connections.

package conn

import (
	"crypto/tls"
	"time"
)

// The connection handler. Implements io.ReadWriteCloser.
type Conn struct {
	// The underlying TLS connection.
	Conn *tls.Conn

	// The timeout duration.
	Timeout time.Duration
}

// Create a new conn object.
func NewConn(conn *tls.Conn, timeout time.Duration) *Conn {
	return &Conn{
		Conn:    conn,
		Timeout: timeout,
	}
}

// Read.
func (c *Conn) Read(p []byte) (n int, err error) {
	// Set the deadline.
	c.Conn.SetDeadline(time.Now().Add(c.Timeout))
	return c.Conn.Read(p)
}

// Write.
func (c *Conn) Write(p []byte) (n int, err error) {
	// Set the deadline.
	c.Conn.SetDeadline(time.Now().Add(c.Timeout))
	return c.Conn.Write(p)
}

// Close.
func (c *Conn) Close() error {
	return c.Conn.Close()
}
