// cmd/deepwell-cli/main.go
// Client CLI program main entrypoint.

package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/cubeflix/deepwell/cli"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const Version = "0.0.0"

var host string
var port int
var skipVerification bool
var key string

// Root command.
func root(cmd *cobra.Command, args []string) {
	if key == "" {
		fmt.Printf("Server Key: ")
		pass, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Println("deepwell-cli:", err.Error())
			os.Exit(1)
		}
		key = string(pass)
	}

	cli := cli.CLI{
		Hostname:         host,
		Addr:             fmt.Sprintf("%s:%d", host, port),
		Key:              key,
		SkipVerification: skipVerification,
	}
	err := cli.Run()
	if err != nil {
		fmt.Println("deepwell-cli:", err.Error())
		os.Exit(1)
	}
}

// Version command.
func version(cmd *cobra.Command, args []string) {
	fmt.Println("deepwell-cli", Version, runtime.GOOS)
}

var rootCmd = &cobra.Command{
	Use:   "deepwell-cli",
	Short: "deepwell-cli is the DEEPWELL command line client",
	Long:  `DEEPWELL is a file server developed by cubeflix at https://github.com/cubeflix/deepwell. deepwell-cli is the command line client program.`,
	Run:   root,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the deepwell-cli version.",
	Run:   version,
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&host, "host", "n", "localhost", "The hostname of the server to connect to. Defaults to localhost.")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 20001, "The port of the server to connect to. Defaults to 20001.")
	rootCmd.PersistentFlags().BoolVarP(&skipVerification, "skip", "s", false, "If the client should skip TLS verification. Defaults to false.")
	rootCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "The access key to use when making requests. If it is not supplied, you will be prompted to input your key.")

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("deepwell-cli:", err.Error())
		os.Exit(1)
	}
}
