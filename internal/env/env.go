package env

import (
	"fmt"
	"os"
	"os/user"

	"golang.org/x/sys/unix"
	"ukiran.com/urm/internal/clog"
)

const MountInfoFile = "/proc/self/mountinfo"

var (
	UID       int
	Username  string
	HomeDir   string
	HomeDevID uint64
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

	var homeStat unix.Stat_t
	if err := unix.Stat(HomeDir, &homeStat); err != nil {
		clog.Error(fmt.Sprintf("failed to Stat %s", HomeDir), "err", err)
		os.Exit(1)
	}

	HomeDevID = homeStat.Dev

	// Extract permission bits (rwxrwxrwx) using a bitwise AND with 0777
	// This yields the standard Unix permission bits.
	// HomePerms = os.FileMode(homeStat.Mode & 0o777)
}

func OpenMountInfo() (*os.File, error) {
	f, err := os.Open(MountInfoFile)
	if err != nil {
		clog.Error("mount file access failed", "file", MountInfoFile, "err", err)
	}
	return f, err
}
