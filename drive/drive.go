// drive/drive.go
// Package drive provides an interface for interacting with DEEPWELL drive
// filesystems.

package drive

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cubeflix/deepwell/protocol"
)

// The drive interface.
type Drive interface {
	// Create a file.
	Create(path string) error

	// Create a directory.
	CreateDirectory(path string) error

	// Read a file into a stream.
	Read(path string, stream io.Writer) error

	// Read a directory.
	ReadDir(path string) ([]os.DirEntry, error)

	// Get information about a file or directory.
	Stat(path string) (os.FileInfo, error)

	// Write a file from a stream.
	Write(path string, stream io.Reader, size int64) error

	// Remove a file or directory. In the case of a directory, the directory
	// must be empty.
	Remove(path string) error

	// Move a file.
	Move(src string, dest string) error
}

// The drive implementation.
type drive struct {
	// The base path of the drive on the host filesystem.
	path string
}

// Create a new drive.
func NewDrive(path string) Drive {
	return &drive{
		path: path,
	}
}

// Convert a local drive path to a path on the host filesystem.
func (d *drive) getHostPath(path string) (string, error) {
	// Clean the path.
	cleanPath := filepath.Clean(path)

	// Check for any "..".
	if strings.Contains(cleanPath, "..") {
		return "", errors.New(fmt.Sprintf("path is invalid: %s", path))
	}

	return filepath.Join(d.path, cleanPath), nil
}

// Create a file.
func (d *drive) Create(path string) error {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return err
	}

	_, err = os.Create(path)
	return err
}

// Create a directory.
func (d *drive) CreateDirectory(path string) error {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return err
	}

	return os.Mkdir(path, 0777)
}

// Read a file into a stream.
func (d *drive) Read(path string, stream io.Writer) error {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return err
	}

	// Open the file.
	file, err := os.OpenFile(path, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file in chunks to the stream.
	reader := bufio.NewReader(file)
	buf := make([]byte, protocol.ChunkSize)
	for {
		// Read the chunk.
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// Write the chunk to the output stream.
		if n < protocol.ChunkSize {
			if _, err := stream.Write(buf[:n]); err != nil {
				return err
			}
			return nil
		}
		if _, err := stream.Write(buf); err != nil {
			return err
		}
	}

	return nil
}

// Read a directory.
func (d *drive) ReadDir(path string) ([]os.DirEntry, error) {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return nil, err
	}

	return os.ReadDir(path)
}

// Get information about a file or directory.
func (d *drive) Stat(path string) (os.FileInfo, error) {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return nil, err
	}

	return os.Stat(path)
}

// Write a file from a stream.
func (d *drive) Write(path string, stream io.Reader, size int64) error {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return err
	}

	// Open the file.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the file in chunks from the stream.
	writer := bufio.NewWriter(file)
	buf := make([]byte, protocol.ChunkSize)
	i := int64(0)
	for {
		var n int
		if size-i < int64(protocol.ChunkSize) {
			// Read the chunk.
			n, err = stream.Read(buf[:size-i])
			if err != nil {
				return err
			}

			// Write the chunk to the file.
			if _, err := writer.Write(buf[:size-i]); err != nil {
				return err
			}
		} else {
			// Read the chunk.
			n, err = stream.Read(buf)
			if err != nil {
				return err
			}

			// Write the chunk to the file.
			if _, err := writer.Write(buf); err != nil {
				return err
			}
		}

		i += int64(n)
		if i >= size {
			break
		}
	}

	// Flush the writer.
	writer.Flush()

	return nil
}

// Remove a file or directory. In the case of a directory, the directory must
// be empty.
func (d *drive) Remove(path string) error {
	// Get the cleaned, final path.
	path, err := d.getHostPath(path)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

// Move a file.
func (d *drive) Move(src string, dest string) error {
	// Get the cleaned, final paths.
	src, err := d.getHostPath(src)
	if err != nil {
		return err
	}
	dest, err = d.getHostPath(dest)
	if err != nil {
		return err
	}

	return os.Rename(src, dest)
}
