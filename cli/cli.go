// cli/cli.go
// The deepwell-cli command-line interface.

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cubeflix/deepwell/client"
	"github.com/google/shlex"
)

// The CLI struct.
type CLI struct {
	Hostname         string
	Addr, Key        string
	SkipVerification bool

	c     client.Client
	drive string
}

// Connect.
func (c *CLI) connect() error {
	c.c = client.NewClient(time.Second * 5)
	c.c.Connect(c.Addr, c.Key)
	c.c.SetInsecureSkipVerify(c.SkipVerification)

	// Ping the server.
	return c.c.Ping()
}

// Run the CLI interface.
func (c *CLI) Run() error {
	fmt.Println("Connecting to", c.Addr)
	err := c.connect()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)

	for {
		// Get the command.
		fmt.Printf("%s:%s> ", c.Hostname, c.drive)
		cmd, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		c.command(cmd)
	}
}

// Perform a command.
func (c *CLI) command(cmd string) {
	args, err := shlex.Split(cmd)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(args) == 0 {
		return
	}
	name := args[0]

	if name == "quit" || name == "exit" {
		// Quit.
		os.Exit(0)
	} else if name == "drive" {
		// Select the drive.
		if len(args) != 2 {
			fmt.Println("Invalid arguments for drive command. Please provide a drive to switch to.")
			return
		}
		c.drive = args[1]
	} else if name == "drives" {
		// Get a list of drives.
		drives, err := c.c.Drives()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(strings.Join(drives, "\n"))
	} else if name == "ping" {
		// Ping the server.
		err := c.c.Ping()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("PONG")
	} else if name == "create" {
		// Create a file.
		if len(args) != 2 {
			fmt.Println("Invalid arguments for create command. Please provide a path to create.")
			return
		}
		if c.drive == "" {
			fmt.Println("No drive selected. Use the drive command to select a drive.")
			return
		}
		err := c.c.Create(c.drive, args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if name == "mkdir" {
		// Create a directory.
		if len(args) != 2 {
			fmt.Println("Invalid arguments for mkdir command. Please provide a path to create.")
			return
		}
		if c.drive == "" {
			fmt.Println("No drive selected. Use the drive command to select a drive.")
			return
		}
		err := c.c.Mkdir(c.drive, args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if name == "download" {
		// Download a file.
		if len(args) != 3 {
			fmt.Println("Invalid arguments for download command. Please provide a path to download and a path to save to.")
			return
		}
		if c.drive == "" {
			fmt.Println("No drive selected. Use the drive command to select a drive.")
			return
		}
		// Open the file for writing.
		f, err := os.Create(args[2])
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}
		n, err := c.c.Read(c.drive, args[1], f)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}
		f.Close()
		fmt.Println("Successfully wrote", n, "bytes to", args[2])
	} else if name == "ls" || name == "list" || name == "dir" {
		// List a directory.
		if len(args) != 1 && len(args) != 2 {
			fmt.Println("Invalid arguments for list command. Please provide a path to list.")
			return
		}
		if c.drive == "" {
			fmt.Println("No drive selected. Use the drive command to select a drive.")
			return
		}
		path := ""
		if len(args) == 2 {
			path = args[1]
		}
		list, err := c.c.List(c.drive, path)
		if err != nil {
			fmt.Println(err)
			return
		}
		for i := range list {
			if list[i].IsDir {
				fmt.Println("D", list[i].Name)
			} else {
				fmt.Println("F", list[i].Name)
			}
		}
	} else if name == "stat" {
		// Stat a path.
		if len(args) != 2 {
			fmt.Println("Invalid arguments for stat command. Please provide a path to stat.")
			return
		}
		if c.drive == "" {
			fmt.Println("No drive selected. Use the drive command to select a drive.")
			return
		}
		stat, err := c.c.Stat(c.drive, args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		if stat.IsDir {
			fmt.Println(args[1])
			fmt.Println("Type: Directory")
		} else {
			fmt.Println(args[1])
			fmt.Println("Type: File")
			fmt.Println("Size:", stat.Size, "bytes")
		}
	} else if name == "upload" {
		// Upload a file.
		if len(args) != 3 {
			fmt.Println("Invalid arguments for upload command. Please provide a path to upload and a path to upload to.")
			return
		}
		if c.drive == "" {
			fmt.Println("No drive selected. Use the drive command to select a drive.")
			return
		}
		// Open the file for reading.
		f, err := os.Open(args[1])
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}
		stat, err := f.Stat()
		if err != nil {
			f.Close()
			return
		}

		// Create the file.
		err = c.c.Create(c.drive, args[2])
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}

		err = c.c.Write(c.drive, args[2], stat.Size(), f)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}
		f.Close()
		fmt.Println("Successfully wrote", stat.Size(), "bytes to", args[2])
	} else {
		fmt.Println("Unrecognized command. Use the 'help' command to get a list of commands.")
	}

	// TODO: help and drive command no drive selected

}
