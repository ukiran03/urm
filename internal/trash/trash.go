package trash

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"ukiran.com/urm/internal/env"
	"ukiran.com/urm/internal/fsys"
)

var (
	ErrNilMinInfo = errors.New("mount-info cannot be nil")
	ErrNoPerms    = fmt.Errorf(
		"have no permissions to write: %w", fs.ErrPermission,
	)
)

type TrashCan struct {
	DevID    uint64 // Device ID of the TopDir on the File System
	TopDir   string // Directory where a file system is mounted
	TrashDir string // Trash location for the TopDir
}

// NewTrashCan returns new TrashCan for the given MountInfo (mountpoint)
func NewTrashCan(minfo *fsys.MountInfo) (*TrashCan, error) {
	if minfo == nil {
		return nil, ErrNilMinInfo
	}
	if minfo.IsReadOnly {
		return nil, ErrNoPerms
	}

	can := &TrashCan{
		DevID:  minfo.DevID,
		TopDir: minfo.MountPoint,
	}

	if minfo.DevID == env.HomeDevID {
		can.TrashDir = getTrashPath(minfo.MountPoint, true)
	} else {
		can.TrashDir = getTrashPath(minfo.MountPoint, false)
	}

	return can, nil
}

func (tc *TrashCan) MkdirP() error {
	return os.MkdirAll(tc.TrashDir, 0o755)
}

// All these methods of TrashCan assume the relevant directories are
// created and checked in prior, see NewTrashCan
func (tc *TrashCan) Move(entry *TrashEntry) error                { return nil }
func (tc *TrashCan) Restore(entry *TrashEntry, dst string) error { return nil }
func (tc *TrashCan) Delete(entry *TrashEntry) error              { return nil }
func (tc *TrashCan) List() ([]*TrashEntry, error)
