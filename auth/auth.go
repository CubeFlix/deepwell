// auth/auth.go
// Package auth provides authentication services to the DEEPWELL server.

package auth

import (
	"errors"
	"fmt"
	"strings"
)

// The authentication interface.
type Authentication interface {
	// Authenticate.
	Authenticate(key, hostname string) (Permissions, error)

	// Add a key.
	AddKey(key string, allowedIPs []string, permissions Permissions)
}

// Authentication implementation.
type authentication struct {
	keys map[string]authKey
}

// An individual authentication key entry.
type authKey struct {
	allowedIPs  map[string]struct{}
	permissions Permissions
}

// Create a new authentication manager.
func NewAuthentication() Authentication {
	return &authentication{keys: map[string]authKey{}}
}

// Authenticate.
func (a *authentication) Authenticate(key, hostname string) (Permissions, error) {
	auth, ok := a.keys[key]
	if !ok {
		return Permissions{}, errors.New(fmt.Sprintf("invalid authentication key: %s", key))
	}

	// Match the hostname.
	hostname = strings.ToLower(hostname)
	if _, ok := auth.allowedIPs[hostname]; !ok {
		return Permissions{}, errors.New(fmt.Sprintf("invalid authentication key: %s", key))
	}

	return auth.permissions, nil
}

// Add a key.
func (a *authentication) AddKey(key string, allowedIPs []string, permissions Permissions) {
	// Create the allowed hosts map.
	hostsMap := map[string]struct{}{}
	for i := range allowedIPs {
		hostsMap[strings.ToLower(allowedIPs[i])] = struct{}{}
	}

	// Create the auth key struct.
	auth := authKey{
		allowedIPs:  hostsMap,
		permissions: permissions,
	}

	a.keys[key] = auth
}
