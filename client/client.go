// client/client.go
// Package client provides an interface for communicating with DEEPWELL
// servers.

package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"time"
)

// The client interface.
type Client interface {
	// Set the address and key of the server to connect to.
	Connect(addr, key string)

	// Insecure skip verify.
	InsecureSkipVerify() bool

	// Set insecure skip verify.
	SetInsecureSkipVerify(v bool)

	// Server name.
	ServerName() string

	// Set server name.
	SetServerName(n string)

	// Add a root CA.
	AddRootCA(cert []byte) error

	// Ping the server.
	Ping() error

	// Get the drives on the server.
	Drives() ([]string, error)

	// Create a file on the server.
	Create(drive, path string) error

	// Create a directory on the server.
	Mkdir(drive, path string) error

	// Read a file on the server into a stream.
	Read(drive, path string, stream io.Writer) (int64, error)

	// List a directory on the server.
	List(drive, path string) ([]DirItem, error)

	// Stat a path on the server.
	Stat(drive, path string) (PathInfo, error)

	// Write a file on the server from a stream. Stops writing once the stream
	// encounters an EOF.
	Write(drive, path string, size int64, stream io.Reader) error

	// Remove a file from the server.
	Remove(drive, path string) error

	// Move a file on the server.
	Move(drive, src, dest string) error
}

// The client implementation.
type client struct {
	addr      string
	key       string
	tlsConfig *tls.Config
	timeout   time.Duration
}

// Create a new client.
func NewClient(timeout time.Duration) Client {
	return &client{tlsConfig: &tls.Config{RootCAs: x509.NewCertPool()}, timeout: timeout}
}

// Insecure skip verify.
func (c *client) InsecureSkipVerify() bool {
	return c.tlsConfig.InsecureSkipVerify
}

// Set insecure skip verify.
func (c *client) SetInsecureSkipVerify(v bool) {
	c.tlsConfig.InsecureSkipVerify = v
}

// Server name.
func (c *client) ServerName() string {
	return c.tlsConfig.ServerName
}

// Set server name.
func (c *client) SetServerName(n string) {
	c.tlsConfig.ServerName = n
}

// Add a root CA.
func (c *client) AddRootCA(cert []byte) error {
	ok := c.tlsConfig.RootCAs.AppendCertsFromPEM(cert)
	if !ok {
		return errors.New("failed to append certificate")
	}
	return nil
}

// Set the address and key of the server to connect to.
func (c *client) Connect(addr, key string) {
	c.addr = addr
	c.key = key
}
