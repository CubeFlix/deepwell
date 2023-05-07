// server/config.go
// Configuration files.

package server

import (
	"crypto/tls"
	"errors"
	"log"
	"os"
	"time"

	"github.com/cubeflix/deepwell/auth"
	"github.com/cubeflix/deepwell/drive"
	"github.com/pelletier/go-toml/v2"
)

// The configuration struct.
type config struct {
	Address          string
	Timeout          string
	Backlog          int
	Workers          int
	SkipVerification bool
	Certificate      []tlsCert
	Logging          logConfig
	Drive            []driveConfig
	Auth             []authConfig
}

// The TLS certificate struct.
type tlsCert struct {
	KeyFile  string
	CertFile string
}

// The logging configuration struct.
type logConfig struct {
	Level string
	File  string
}

// The drive configuration struct.
type driveConfig struct {
	Name string
	Path string
}

// The authentication configuration struct.
type authConfig struct {
	Key           string
	AllowedIPs    []string
	AllowedDrives []string
	CanWrite      bool
}

// Empty writer.
type emptyWriter struct{}

func (w *emptyWriter) Write(b []byte) (n int, err error) {
	return len(b), nil
}

// Load a configuration file.
func (s *server) LoadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Load the TOML file.
	var cfg config = config{
		Address:          ":20001",
		Timeout:          "3s",
		Backlog:          10,
		Workers:          5,
		SkipVerification: false,
		Certificate:      []tlsCert{},
		Logging:          logConfig{},
		Drive:            []driveConfig{},
		Auth:             []authConfig{},
	}
	err = toml.Unmarshal(file, &cfg)
	if err != nil {
		return err
	}

	s.SetAddress(cfg.Address)
	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return err
	}
	s.SetTimeout(timeout)
	s.SetBacklogSize(cfg.Backlog)
	s.SetNumWorkers(cfg.Workers)

	// load the drives.
	drives := map[string]drive.Drive{}
	for i := range cfg.Drive {
		if cfg.Drive[i].Name == "" || cfg.Drive[i].Path == "" {
			return errors.New("drive configuration must contain name and path")
		}
		drives[cfg.Drive[i].Name] = drive.NewDrive(cfg.Drive[i].Path)
	}
	s.SetDrives(drives)

	// Load the authentication.
	authentication := auth.NewAuthentication()
	for i := range cfg.Auth {
		if cfg.Auth[i].Key == "" || cfg.Auth[i].AllowedDrives == nil || cfg.Auth[i].AllowedIPs == nil {
			return errors.New("auth configuration must contain key, allowed drives, and allowed IPs")
		}
		authentication.AddKey(cfg.Auth[i].Key, cfg.Auth[i].AllowedIPs, auth.Permissions{AllowedDrives: cfg.Auth[i].AllowedDrives, CanWrite: cfg.Auth[i].CanWrite})
	}
	s.SetAuthentication(authentication)

	// Load the TLS configuration.
	certs := []tls.Certificate{}
	for i := range cfg.Certificate {
		cert, err := tls.LoadX509KeyPair(cfg.Certificate[i].CertFile, cfg.Certificate[i].KeyFile)
		if err != nil {
			return err
		}
		certs = append(certs, cert)
	}
	s.SetTLSConfig(&tls.Config{
		Certificates:       certs,
		InsecureSkipVerify: cfg.SkipVerification,
	})

	// Load the logger.
	logFile := os.Stdout
	if cfg.Logging.File != "" {
		logFile, err = os.OpenFile(cfg.Logging.File, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		s.logFile = logFile
	}

	if cfg.Logging.Level == "none" {
		s.info = log.New(&emptyWriter{}, "info: ", log.Ldate|log.Ltime|log.Lshortfile)
		s.err = log.New(&emptyWriter{}, "error: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else if cfg.Logging.Level == "info" {
		s.info = log.New(logFile, "info: ", log.Ldate|log.Ltime|log.Lshortfile)
		s.err = log.New(logFile, "error: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else if cfg.Logging.Level == "error" {
		s.info = log.New(&emptyWriter{}, "info: ", log.Ldate|log.Ltime|log.Lshortfile)
		s.err = log.New(logFile, "error: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		s.info = log.New(logFile, "info: ", log.Ldate|log.Ltime|log.Lshortfile)
		s.err = log.New(logFile, "error: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	return nil
}
