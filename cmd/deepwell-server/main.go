// cmd/deepwell-server/main.go
// Server program main entrypoint.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cubeflix/deepwell/server"
	"github.com/spf13/cobra"
)

const Version = "0.0.0"

var cfgFile string

// Version command.
func version(cmd *cobra.Command, args []string) {
	fmt.Println("deepwell-server", Version, runtime.GOOS)
}

// Serve command.
func serve(cmd *cobra.Command, args []string) {
	if cfgFile == "" {
		cfgFile = ".deepwell.toml"
	}

	// Create the server.
	s := server.NewServer()
	err := s.LoadConfig(cfgFile)
	if err != nil {
		fmt.Println("deepwell-server:", err.Error())
		os.Exit(1)
	}
	go s.Serve()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt)
	<-stop
	s.Stop()
}

var rootCmd = &cobra.Command{
	Use:   "deepwell-server",
	Short: "deepwell-server is the DEEPWELL file server program",
	Long:  `DEEPWELL is a file server developed by cubeflix at https://github.com/cubeflix/deepwell. deepwell-server is the main server program.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the deepwell-server version.",
	Run:   version,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start serving the DEEPWELL server.",
	Run:   serve,
}

func main() {
	serveCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "The server config TOML file. Defaults to .deepwell.toml.")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("deepwell-server:", err.Error())
		os.Exit(1)
	}
}
