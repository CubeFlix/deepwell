// auth/permissions.go
// Provides permissions management for authentication.

package auth

// Permissions struct.
type Permissions struct {
	AllowedDrives []string
	CanWrite      bool
}

func (p *Permissions) DriveAllowed(drive string) bool {
	for _, a := range p.AllowedDrives {
		if a == drive {
			return true
		}
	}
	return false
}
