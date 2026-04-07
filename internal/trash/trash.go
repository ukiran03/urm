package trash

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
	"ukiran.com/useme/internal/env"
	"ukiran.com/useme/internal/fsys"
)

type TrashCan struct {
	TopDir   string // Directory where a file system is mounted
	DevID    uint64 // Device ID of the TopDir on the File System
	TrashDir string // Trash location for the TopDir
}

type CanType int

const (
	HomeCan    CanType = iota
	SpecialCan         // Non-Home can
)

// NewTrashCan returns a new TrashCan for the given MountInfo (e.g.,
// partition or device).  It does not initialize the TrashCan for
// immediate operations by default.
func NewTrashCan(minfo *fsys.MountInfo, canType CanType) (*TrashCan, error) {
	if minfo == nil {
		return nil, fmt.Errorf("ERR: mount info cannot be nil")
	}
	var trashPath string
	var err error

	switch canType {
	case HomeCan:
		var homeStat unix.Stat_t
		if err = unix.Stat(env.HomeDir, &homeStat); err != nil {
			return nil, &TrashError{Op: "Stat", Path: env.HomeDir, Err: err}
		}
		// Ensure we are on the same device
		if homeStat.Dev != minfo.DevID {
			return nil, fmt.Errorf("ERR: device ID mismatch for home directory")
		}
		trashPath, err = getHomeTrashPath(env.HomeDir)
		if err != nil {
			return nil, &TrashError{Op: "getHomeTrashPath", Path: env.HomeDir, Err: err}
		}

	case SpecialCan:
		// Ensure MountPoint isn't empty
		if minfo.MountPoint == "" {
			return nil, errors.New("ERR: Empty MountPoint for special can")
		}
		trashPath, err = getSpecialTrashPath(minfo.MountPoint, env.UID)
		if err != nil {
			return nil, &TrashError{
				Op:   "getSpecialTrashDir",
				Path: minfo.MountPoint, Err: err,
			}
		}
	default:
		return nil, fmt.Errorf("ERR: unsupported trash can type: %v", canType)
	}

	return &TrashCan{
		TopDir:   minfo.MountPoint,
		DevID:    minfo.DevID,
		TrashDir: trashPath,
	}, nil
}

// All these methods of TrashCan assume the relevant directories are
// created and checked in prior, see NewTrashCan
func (tc *TrashCan) Move(entry *FileEntry) error                { return nil }
func (tc *TrashCan) Restore(entry *FileEntry, dst string) error { return nil }
func (tc *TrashCan) Delete(entry *FileEntry) error              { return nil }
func (tc *TrashCan) List() ([]*FileEntry, error)
