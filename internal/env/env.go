package env

import (
	"os"
	"os/user"
)

var (
	UID      int
	Username string
	HomeDir  string
)

// Init initializes the global environment variables.
// Call this once at the very start of your main() function.
func Init() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	UID = os.Getuid()
	Username = u.Username
	HomeDir = u.HomeDir

	return nil
}
