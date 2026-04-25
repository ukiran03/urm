package trash

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// TrashEntry is the representation of a file to be trashed
type TrashEntry struct {
	OrigPath string // AbsPath where the file lived before trashing

	// TODO: remove one field
	TrashPath     string // AbsPath to the actual data in $topdir/files
	TrashInfoPath string // AbsPath to the corresponding .trashinfo

	MountRoot    string // Root of filesystem where the file originally resided
	Name         string // Base name of the file
	Size         int64  // Logical size of file in bytes
	IsDir        bool
	FileMode     fs.FileMode
	DeletionDate time.Time    // Time of Trashing operation
	Stat         *unix.Stat_t // Raw Unix syscall metadata (Inodes, Device IDs)
}

// NewTrashEntry initializes a TrashEntry struct using the file's
// absolute path and its os.FileInfo (obtained via os.Lstat). It
// extracts underlying Unix syscall metadata; returns a non-nil error
// if the system metadata is inaccessible.
func NewTrashEntry(absPath string, info os.FileInfo) (*TrashEntry, error) {
	stat, ok := info.Sys().(*unix.Stat_t)
	if !ok {
		return nil, fmt.Errorf(
			"failed to get unix syscall.Stat_t for %s", absPath,
		)
	}
	// TrashPath, TrashInfoPath, MountRoot will be added later by
	// other functions
	entry := &TrashEntry{
		OrigPath:     absPath,
		Name:         info.Name(),
		Size:         info.Size(),
		IsDir:        info.IsDir(),
		FileMode:     info.Mode(),
		DeletionDate: time.Now().UTC(), // To comply with most metadata standards
		Stat:         stat,
	}
	return entry, nil
}

// SetTrashPath is the "second pass" function.
func (f *TrashEntry) SetTrashPath(destination string) {
	f.TrashPath = destination
}

func (f *TrashEntry) DirSize() (int64, error) {
	// TODO: fsys.ConcurrnetDirSize(f.OrigPath)
	return 0, nil
}
