// client/commands.go
// Server commands.

package client

import (
	"errors"
	"io"
	"strconv"

	"github.com/cubeflix/deepwell/protocol"
)

// Ping the server.
func (c *client) Ping() error {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("ping", c.key, "")
	if err != nil {
		return err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return err
	}

	// Receive the "PONG" response.
	_, err = r.getString()
	if err != nil {
		return err
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return err
	}

	return nil
}

// Get the drives on the server.
func (c *client) Drives() ([]string, error) {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return nil, err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("drives", c.key, "")
	if err != nil {
		return nil, err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return nil, err
	}

	// Receive the number of drives.
	numDrivesStr, err := r.getString()
	if err != nil {
		return nil, err
	}
	numDrives, err := strconv.Atoi(numDrivesStr)
	if err != nil {
		return nil, err
	}

	drives := make([]string, numDrives)
	for i := range drives {
		driveName, err := r.getString()
		if err != nil {
			return nil, err
		}
		drives[i] = driveName
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return nil, err
	}

	return drives, nil
}

// Create a file on the server.
func (c *client) Create(drive, path string) error {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("create", c.key, drive+"\n"+path+"\n")
	if err != nil {
		return err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return err
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return err
	}

	return nil
}

// Create a directory on the server.
func (c *client) Mkdir(drive, path string) error {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("mkdir", c.key, drive+"\n"+path+"\n")
	if err != nil {
		return err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return err
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return err
	}

	return nil
}

// Read a file on the server into a stream.
func (c *client) Read(drive, path string, stream io.Writer) (int64, error) {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return 0, err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("read", c.key, drive+"\n"+path+"\n")
	if err != nil {
		return 0, err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return 0, err
	}

	// Get the length of the data.
	lenStr, err := r.getString()
	if err != nil {
		return 0, err
	}
	_, err = strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return io.Copy(stream, r.reader)
}

// A directory list item.
type DirItem struct {
	Name  string
	IsDir bool
}

// List a directory on the server.
func (c *client) List(drive, path string) ([]DirItem, error) {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return nil, err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("list", c.key, drive+"\n"+path+"\n")
	if err != nil {
		return nil, err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return nil, err
	}

	// Receive the number of items.
	numItemsStr, err := r.getString()
	if err != nil {
		return nil, err
	}
	numItems, err := strconv.Atoi(numItemsStr)
	if err != nil {
		return nil, err
	}

	items := make([]DirItem, numItems)
	for i := range items {
		line, err := r.getString()
		if err != nil {
			return nil, err
		}
		if len(line) < 2 {
			return nil, errors.New("invalid server response")
		}
		if line[0] == 'd' {
			items[i].IsDir = true
		}
		items[i].Name = line[2:]
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return nil, err
	}

	return items, nil
}

// Path information.
type PathInfo struct {
	IsDir bool
	Size  int64
}

// Stat a path on the server.
func (c *client) Stat(drive, path string) (PathInfo, error) {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return PathInfo{}, err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("stat", c.key, drive+"\n"+path+"\n")
	if err != nil {
		return PathInfo{}, err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return PathInfo{}, err
	}

	// Receive the number of items.
	line, err := r.getString()
	if err != nil {
		return PathInfo{}, err
	}

	info := PathInfo{}

	if line[0] == 'd' {
		info.IsDir = true
	} else {
		if len(line) < 2 {
			return PathInfo{}, errors.New("invalid server response")
		}
		sizeStr := line[2:]
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return PathInfo{}, err
		}
		info.Size = size
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return PathInfo{}, err
	}

	return info, nil
}

// Write a file on the server from a stream. Stops writing once the stream
// encounters an EOF.
func (c *client) Write(drive, path string, size int64, stream io.Reader) error {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return err
	}
	defer r.conn.Close()

	// Send the header.
	err = r.sendString(protocol.Header)
	if err != nil {
		return err
	}

	// Send the key and command.
	err = r.sendString(c.key)
	if err != nil {
		return err
	}
	err = r.sendString("write")
	if err != nil {
		return err
	}

	data := drive + "\n" + path + "\n"

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

	// Send the length of the data.
	err = r.sendString(strconv.FormatInt(size, 10))
	if err != nil {
		return err
	}

	// Send the data.
	_, err = io.Copy(r.writer, stream)
	if err != nil {
		return err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return err
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return err
	}

	return nil
}

// Remove a file from the server.
func (c *client) Remove(drive, path string) error {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("remove", c.key, drive+"\n"+path+"\n")
	if err != nil {
		return err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return err
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return err
	}

	return nil
}

// Move a file on the server.
func (c *client) Move(drive, src, dest string) error {
	// Create a connection.
	r, err := c.newRequest()
	if err != nil {
		return err
	}
	defer r.conn.Close()

	// Send the request.
	err = r.sendSimpleRequest("move", c.key, drive+"\n"+src+"\n"+dest+"\n")
	if err != nil {
		return err
	}

	// Receive the header.
	err = r.receiveHeader()
	if err != nil {
		return err
	}

	// Consume.
	err = r.consume()
	if err != nil {
		return err
	}

	return nil
}
