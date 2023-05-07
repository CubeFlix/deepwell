// client/request.go
// Server requests.

package client

import (
	"bufio"
	"crypto/tls"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/cubeflix/deepwell/conn"
	"github.com/cubeflix/deepwell/protocol"
)

// The request struct.
type request struct {
	// The underlying connection. The reader and writer should be used in all
	// cases.
	conn   *tls.Conn
	writer *conn.Conn
	reader *bufio.Reader

	// The request information.
	command string
}

// Create a new request.
func newRequest(c *tls.Conn, timeout time.Duration) *request {
	conn := conn.NewConn(c, timeout)
	return &request{
		conn:   c,
		writer: conn,
		reader: bufio.NewReader(conn),
	}
}

// Create a new request.
func (c *client) newRequest() (*request, error) {
	conn, err := tls.Dial("tcp", c.addr, c.tlsConfig)
	if err != nil {
		return nil, err
	}
	return newRequest(conn, c.timeout), nil
}

// Get a string from the connection. Terminates once it reaches a newline.
func (r *request) getString() (string, error) {
	// Scan the string.
	str, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return str[:len(str)-1], nil
}

// Send a string over the connection.
func (r *request) sendString(s string) error {
	// Send the string, along with a newline.
	_, err := r.writer.Write([]byte(s + "\n"))
	return err
}

// Send a simple request (does not require chunk data).
func (r *request) sendSimpleRequest(command, key, data string) error {
	// Send the header.
	err := r.sendString(protocol.Header)
	if err != nil {
		return err
	}

	// Send the key and command.
	err = r.sendString(key)
	if err != nil {
		return err
	}
	err = r.sendString(command)
	if err != nil {
		return err
	}

	// Send the length of the data.
	err = r.sendString(strconv.Itoa(len(data)))
	if err != nil {
		return err
	}

	// Send the data.
	_, err = r.writer.Write([]byte(data))
	if err != nil {
		return err
	}

	return r.sendString("0")
}

// Receive the header.
func (r *request) receiveHeader() error {
	// Receive the header.
	header, err := r.getString()
	if err != nil {
		return err
	}
	if header != protocol.Header {
		return errors.New("invalid header")
	}

	// Receive the status.
	status, err := r.getString()
	if err != nil {
		return err
	}
	if strings.ToLower(status) == "failed" {
		// Failed. Receive the error.
		errString, err := r.getString()
		if err != nil {
			return err
		}
		return errors.New(errString)
	}
	if strings.ToLower(status) != "success" {
		return errors.New("invalid status response")
	}

	return nil
}

// Consume a chunk of data, prefixed with the length.
func (r *request) consume() error {
	// Get the length of the data.
	lenStr, err := r.getString()
	if err != nil {
		return err
	}
	len, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		return err
	}

	buf := make([]byte, protocol.ChunkSize)
	n := int64(0)
	for {
		// Read the chunk.
		if len-n < int64(protocol.ChunkSize) {
			smallBuf := make([]byte, len-n)
			_, err := r.reader.Read(smallBuf)
			if err != nil {
				return err
			}
			return nil
		} else {
			_, err := r.reader.Read(buf)
			if err != nil {
				return err
			}
		}
	}
}
