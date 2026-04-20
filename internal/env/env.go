package env

import (
	"os"
	"os/user"

	"ukiran.com/urm/internal/clog"
)

var (
	UID           int
	Username      string
	HomeDir       string
	MountInfoFile = "/proc/self/mountinfo"
)

func init() {
	u, err := user.Current()
	if err != nil {
		clog.Error("failed to retrieve current user context", "err", err)
		os.Exit(1)
	}

	UID = os.Getuid()
	Username = u.Username
	HomeDir = u.HomeDir
}

func OpenMountInfo() (*os.File, error) {
	f, err := os.Open(MountInfoFile)
	if err != nil {
		clog.Error("mount file access failed", "file", MountInfoFile, "err", err)
	}
	return f, err
}
