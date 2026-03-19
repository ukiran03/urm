package trash

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
	"ukiran.com/useme/internal/fsys"
)

// getMountsInfos: reads and parses mountinfoFile and produces slice
// of available trash paths(ie $topdirs)
func getMountsInfos(mountinfofile string) ([]*fsys.MountInfo, error) {
	f, err := os.Open(mountinfofile)
	if err != nil {
		return nil, fmt.Errorf("Error opening mountinfo file: %w", err)
	}
	mounts, err := fsys.ParseMountInfo(f, fsys.IgnoreFsFunc)
	if err != nil {
		return nil, fmt.Errorf("Error parsing mountinfo file: %w", err)
	}
	return mounts, nil
}

// onSameDevice: determines that the given two devIds are equal, hence
// both belong to the same device
func onSameDevice(dstDevId, srcDevId uint64) bool {
	if dstDevId == srcDevId {
		return true
	}
	return false
}

func dirExists(path string) (os.FileInfo, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path explicitly does not exist
			return nil, false, nil
		}
		// Some other error (Permissions, etc.)
		return nil, false, err
	}
	// No error, so the path exists. Now check if it's a directory.
	return info, info.IsDir(), nil
}

// Ensure PATH is secure (Sticky Bit).
// os.ModeSticky is 0x20000000; in chmod terms, it's the "1" in "1777"
func haveSticyBit(path string, info os.FileInfo) (bool, error) {
	if (info.Mode() & os.ModeSticky) == 0 {
		return false, fmt.Errorf(
			// other users could delete files
			"security risk: sticky bit not set on %s", path,
		)
	}
	return true, nil
}

// Check ownership (Standard requirement: must be owned by root and global-writable)
func havePermissions(path string, info os.FileInfo) (bool, error) {
	stat, ok := info.Sys().(*unix.Stat_t)
	if !ok {
		return false, fmt.Errorf(
			"could not get raw unix.Stat_t for %s", path)
	}
	// Ownership check (must be root)
	if stat.Uid != 0 {
		return false, fmt.Errorf(
			"security risk: %s is owned by UID %d, must be root (0)",
			path, stat.Uid,
		)
	}

	// Writable check
	// For a public trash dir, we usually want 0777 or 0775.  If it's
	// 0700 and owned by root, user won't be able to create their
	// $uid folder.
	mode := info.Mode().Perm()
	if (mode & 0o02) == 0 {
		return false, fmt.Errorf(
			"directory %s is not global-writable, user cannot create trash subfolders",
			path,
		)
	}

	return true, nil
}
