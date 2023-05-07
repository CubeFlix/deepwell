// server/request.go
// Handles requests to the DEEPWELL server.

package server

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cubeflix/deepwell/auth"
	"github.com/cubeflix/deepwell/conn"
	"github.com/cubeflix/deepwell/drive"
	"github.com/cubeflix/deepwell/protocol"
)

// The request struct.
type request struct {
	// The underlying connection. The reader and writer should be used in all
	// cases.
	conn   *tls.Conn
	writer *conn.Conn
	reader *bufio.Reader

	// Authentication information.
	key         string
	permissions auth.Permissions

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

// Handle a single request.
func (s *server) handleRequest(r *request) error {
	defer r.conn.Close()

	// Read the DEEPWELL protocol header.
	header, err := r.getString()
	if err != nil {
		return err
	}
	if header != protocol.Header {
		// Close the connection, we got an invalid header.
		return nil
	}

	// Read the authentication information.
	key, err := r.getString()
	if err != nil {
		return err
	}
	r.key = key

	// Read the command.
	command, err := r.getString()
	if err != nil {
		return err
	}
	command = strings.ToLower(command)
	r.command = command

	// Authenticate the user.
	ip, _, err := net.SplitHostPort(r.conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	permissions, err := s.authentication.Authenticate(key, ip)
	if err != nil {
		s.info.Println("failed to authenticate user:", key, ip)
		// Failed to log in.
		if err := r.consume(); err != nil {
			return err
		}
		if err := r.consume(); err != nil {
			return err
		}
		if err := r.sendError(err.Error()); err != nil {
			return err
		}
		return nil
	}
	r.permissions = permissions

	// Handle the command.
	function, ok := s.commands[command]
	if !ok {
		// Invalid command.
		if err := r.consume(); err != nil {
			return err
		}
		if err := r.consume(); err != nil {
			return err
		}
		if err := r.sendError(fmt.Sprintf("invalid command %s", command)); err != nil {
			return err
		}
		return nil
	}

	// Invoke the command. It is up to the command to handle responses/errors.
	err = function(r)
	return err
}

// Get a drive, given a server.
func (r *request) getDrive(drive string, s Server) (drive.Drive, error) {
	// Check if the user can access the drive.
	ok := r.permissions.DriveAllowed(drive)
	if !ok {
		// Not allowed.
		return nil, errors.New(fmt.Sprintf("drive not allowed: %s", drive))
	}

	// Attempt to find the drive.
	driveObj, ok := s.Drives()[drive]
	if !ok {
		// Drive does not exist.
		return nil, errors.New(fmt.Sprintf("drive not allowed: %s", drive))
	}

	return driveObj, nil
}

// Send an error response.
func (r *request) sendError(s string) error {
	if err := r.sendString(protocol.Header); err != nil {
		return err
	}
	if err := r.sendString("FAILED"); err != nil {
		return err
	}
	if err := r.sendString(s); err != nil {
		return err
	}
	if err := r.sendString("0"); err != nil {
		return err
	}
	return nil
}

// Send an simple success response.
func (r *request) sendSuccess(s string) error {
	if err := r.sendString(protocol.Header); err != nil {
		return err
	}
	if err := r.sendString("SUCCESS"); err != nil {
		return err
	}
	if _, err := r.writer.Write([]byte(s)); err != nil {
		return err
	}
	if err := r.sendString("0"); err != nil {
		return err
	}
	return nil
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
			i, err := r.reader.Read(buf)
			if err != nil {
				return err
			}
			n += int64(i)
		}
	}
}
