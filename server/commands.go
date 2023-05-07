// server/commands.go
// Implementation of functions/commands on the DEEPWELL server.

package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cubeflix/deepwell/protocol"
)

// Ping command.
func (s *server) pingCommand(r *request) error {
	// Consume.
	if err := r.consume(); err != nil {
		return err
	}
	if err := r.consume(); err != nil {
		return err
	}

	return r.sendSuccess("PONG\n")
}

// Drives command.
func (s *server) drivesCommand(r *request) error {
	// Consume.
	if err := r.consume(); err != nil {
		return err
	}
	if err := r.consume(); err != nil {
		return err
	}

	numDrivesStr := strconv.Itoa(len(r.permissions.AllowedDrives))
	return r.sendSuccess(numDrivesStr + "\n" + strings.Join(r.permissions.AllowedDrives, "\n") + "\n")
}

// Create command.
func (s *server) createCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path of the file to create.
	path, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	if !r.permissions.CanWrite {
		err := r.sendError("no write permissions")
		if err != nil {
			return err
		}
		return nil
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	// Attempt to create the file.
	err = drive.Create(path)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	s.info.Println("create", path)

	return r.sendSuccess("")
}

// Mkdir command.
func (s *server) mkdirCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path of the directory to create.
	path, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	if !r.permissions.CanWrite {
		err := r.sendError("no write permissions")
		if err != nil {
			return err
		}
		return nil
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	// Attempt to create the directory.
	err = drive.CreateDirectory(path)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	s.info.Println("mkdir", path)

	return r.sendSuccess("")
}

// Read command.
func (s *server) readCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path of the file to read.
	path, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	// Get the size of the data and ensure it is a file.
	stat, err := drive.Stat(path)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}
	if stat.IsDir() {
		err = r.sendError(fmt.Sprintf("cannot be read: %s", path))
		if err != nil {
			return err
		}
		return nil
	}

	s.info.Println("read", path)

	if err := r.sendString(protocol.Header); err != nil {
		return err
	}
	if err := r.sendString("SUCCESS"); err != nil {
		return err
	}
	if err := r.sendString(strconv.FormatInt(stat.Size(), 10)); err != nil {
		return err
	}
	return drive.Read(path, r.writer)
}

// List directory command.
func (s *server) listCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path of the directory to list.
	path, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	items, err := drive.ReadDir(path)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}
	numItemsStr := strconv.Itoa(len(items))
	text := ""
	for i := range items {
		if items[i].IsDir() {
			text += "d " + items[i].Name() + "\n"
		} else {
			text += "f " + items[i].Name() + "\n"
		}
	}

	s.info.Println("list", path)

	return r.sendSuccess(numItemsStr + "\n" + text)
}

// Stat command.
func (s *server) statCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path.
	path, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	stat, err := drive.Stat(path)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	s.info.Println("stat", path)

	if stat.IsDir() {
		return r.sendSuccess("d\n")
	} else {
		return r.sendSuccess("f " + strconv.FormatInt(stat.Size(), 10) + "\n")
	}
}

// Write command.
func (s *server) writeCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path of the file to write.
	path, err := r.getString()
	if err != nil {
		return err
	}

	if !r.permissions.CanWrite {
		// Consume.
		err2 := r.consume()
		if err2 != nil {
			return err2
		}

		err := r.sendError("no write permissions")
		if err != nil {
			return err
		}
		return nil
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		// Consume.
		err2 := r.consume()
		if err2 != nil {
			return err2
		}

		err2 = r.sendError(err.Error())
		if err2 != nil {
			return err2
		}
		return nil
	}

	// Ensure it is a file.
	stat, err := drive.Stat(path)
	if err != nil {
		// Consume.
		err2 := r.consume()
		if err2 != nil {
			return err2
		}

		err2 = r.sendError(err.Error())
		if err2 != nil {
			return err2
		}
		return nil
	}
	if stat.IsDir() {
		// Consume.
		err = r.consume()
		if err != nil {
			return err
		}

		err = r.sendError(fmt.Sprintf("cannot be read: %s", path))
		if err != nil {
			return err
		}
		return nil
	}

	// Read the size of the data.
	lenStr, err := r.getString()
	if err != nil {
		return err
	}
	len, err := strconv.ParseInt(lenStr, 0, 64)
	if err != nil {
		return err
	}

	// Write
	if err := drive.Write(path, r.reader, len); err != nil {
		return err
	}

	s.info.Println("write", path)

	return r.sendSuccess("")
}

// Remove command.
func (s *server) removeCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the path to remove.
	path, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	if !r.permissions.CanWrite {
		err := r.sendError("no write permissions")
		if err != nil {
			return err
		}
		return nil
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	// Attempt to remove the path.
	err = drive.Remove(path)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	s.info.Println("remove", path)

	return r.sendSuccess("")
}

// Move command.
func (s *server) moveCommand(r *request) error {
	if _, err := r.getString(); err != nil {
		return err
	}

	// Get the drive.
	driveName, err := r.getString()
	if err != nil {
		return err
	}

	// Get the source path.
	src, err := r.getString()
	if err != nil {
		return err
	}

	// Get the destination path.
	dest, err := r.getString()
	if err != nil {
		return err
	}

	// Consume.
	if err := r.consume(); err != nil {
		return err
	}

	if !r.permissions.CanWrite {
		err := r.sendError("no write permissions")
		if err != nil {
			return err
		}
		return nil
	}

	// Get the drive.
	drive, err := r.getDrive(driveName, s)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	// Attempt to move the paths.
	err = drive.Move(src, dest)
	if err != nil {
		err = r.sendError(err.Error())
		if err != nil {
			return err
		}
		return nil
	}

	s.info.Println("move", src, dest)

	return r.sendSuccess("")
}
