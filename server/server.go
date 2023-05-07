// server/server.go
// Package server provides an interface for creating and running DEEPWELL
// servers.

package server

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"time"

	"github.com/cubeflix/deepwell/auth"
	"github.com/cubeflix/deepwell/drive"
)

// The server interface.
type Server interface {
	// Get the server address.
	Address() string

	// Set the server address.
	SetAddress(addr string)

	// Get the timeout duration.
	Timeout() time.Duration

	// Set the timeout duration.
	SetTimeout(timeout time.Duration)

	// TLS config.
	TLSConfig() *tls.Config

	// Set the TLS config.
	SetTLSConfig(config *tls.Config)

	// Get the backlog size.
	BacklogSize() int

	// Set the backlog size.
	SetBacklogSize(size int)

	// Get the number of workers.
	NumWorkers() int

	// Set the number of workers.
	SetNumWorkers(workers int)

	// Get the map of drives.
	Drives() map[string]drive.Drive

	// Set the map of drives.
	SetDrives(drives map[string]drive.Drive)

	// Get the loggers.
	Logger() (info, err *log.Logger)

	// Set the loggers.
	SetLogger(info, err *log.Logger)

	// Get the authentication manager.
	Authentication() auth.Authentication

	// Set the authentication manager.
	SetAuthentication(auth auth.Authentication)

	// Load a configuration file.
	LoadConfig(path string) error

	// Serve.
	Serve() error

	// Stop serving.
	Stop()
}

// The server implementation.
type server struct {
	addr           string
	timeout        time.Duration
	tlsConfig      *tls.Config
	backlogSize    int
	numWorkers     int
	drives         map[string]drive.Drive
	authentication auth.Authentication

	info    *log.Logger
	err     *log.Logger
	logFile *os.File

	commands map[string]func(*request) error

	running    bool
	jobs       chan *request
	stopSignal chan struct{}
	listener   net.Listener
}

// Create a new server.
func NewServer() Server {
	s := &server{authentication: auth.NewAuthentication()}
	s.commands = map[string]func(*request) error{
		"ping":   s.pingCommand,
		"drives": s.drivesCommand,
		"create": s.createCommand,
		"mkdir":  s.mkdirCommand,
		"read":   s.readCommand,
		"list":   s.listCommand,
		"stat":   s.statCommand,
		"write":  s.writeCommand,
		"remove": s.removeCommand,
		"move":   s.moveCommand,
	}
	return s
}

// Get the server address.
func (s *server) Address() string {
	return s.addr
}

// Set the server address.
func (s *server) SetAddress(addr string) {
	s.addr = addr
}

// Get the timeout duration.
func (s *server) Timeout() time.Duration {
	return s.timeout
}

// Set the timeout duration.
func (s *server) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// Get the TLS config.
func (s *server) TLSConfig() *tls.Config {
	return s.tlsConfig
}

// Set the TLS config.
func (s *server) SetTLSConfig(config *tls.Config) {
	s.tlsConfig = config
}

// Get the backlog size.
func (s *server) BacklogSize() int {
	return s.backlogSize
}

// Set the backlog size.
func (s *server) SetBacklogSize(size int) {
	s.backlogSize = size
}

// Get the number of workers.
func (s *server) NumWorkers() int {
	return s.numWorkers
}

// Set the number of workers.
func (s *server) SetNumWorkers(workers int) {
	s.numWorkers = workers
}

// Get the map of drives.
func (s *server) Drives() map[string]drive.Drive {
	return s.drives
}

// Set the map of drives.
func (s *server) SetDrives(drives map[string]drive.Drive) {
	s.drives = drives
}

// Get the loggers.
func (s *server) Logger() (info, err *log.Logger) {
	return s.info, s.err
}

// Set the loggers.
func (s *server) SetLogger(info, err *log.Logger) {
	s.info = info
	s.err = err
}

// Get the authentication manager.
func (s *server) Authentication() auth.Authentication {
	return s.authentication
}

// Set the authentication manager.
func (s *server) SetAuthentication(a auth.Authentication) {
	s.authentication = a
}

// Serve.
func (s *server) Serve() error {
	s.running = true

	// Initialize the channels.
	s.jobs = make(chan *request, s.backlogSize)
	s.stopSignal = make(chan struct{}, s.numWorkers)

	// Start the workers.
	for i := 0; i < s.numWorkers; i++ {
		go s.worker()
	}

	s.info.Println("starting server")

	// Start listening.
	return s.listen()
}

// Stop serving.
func (s *server) Stop() {
	// Stop listening.
	s.running = false
	s.listener.Close()

	// Stop the workers.
	for i := 0; i < s.numWorkers; i++ {
		s.stopSignal <- struct{}{}
	}

	s.info.Println("stopping server")

	// Close the log file.
	if s.logFile != nil {
		s.logFile.Close()
	}
}

// The connection handling routine.
func (s *server) listen() error {
	// Create the listener.
	listener, err := tls.Listen("tcp", s.addr, s.tlsConfig)
	s.listener = listener
	if err != nil {
		return err
	}

	// Accept connections.
	for s.running {
		conn, err := listener.Accept()
		if err != nil {
			if !s.running {
				// If we are not running (i.e. shutting down), then ignore this
				// and exit.
				return nil
			}
			s.err.Println("failed to accept connection: ", err.Error())
			continue
		}
		req := newRequest(conn.(*tls.Conn), s.timeout)
		s.jobs <- req
	}

	return nil
}

// The worker routine.
func (s *server) worker() error {
	// Continually handle new requests.
	for s.running {
		select {
		case <-s.stopSignal:
			// Stop signal. NOTE: Never put any code here since we can't be
			// sure we'll ever get the stop signal, we may just exit the loop.
			return nil
		case req := <-s.jobs:
			// Got a request.
			if err := s.handleRequest(req); err != nil {
				s.err.Println("failed to handle request: ", err.Error())
			}
		}
	}

	return nil
}
