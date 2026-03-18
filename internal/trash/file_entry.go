package trash

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// FileEntry is the representation of a file in trash
type FileEntry struct {
	OrigPath     string      // Absolute path of file "prior" trashing
	Name         string      // Original base name of file
	Size         int64       // Size of file in bytes
	IsDir        bool        // Indicates if this is a directory
	FileMode     fs.FileMode // Mode of the file
	DeletionDate time.Time   // Time of deletion
	MountRoot    string      // Root path of the Mount Point
	DeviceID     uint64      // File's Device ID
	TrashPath    string      // Absolute path of file "after" trashing
}

// NewFileEntry takes in the Absoute path of the file (must exist) and
// returns the FileEntry
func NewFileEntry(filePath string) (*FileEntry, error) {
	info, err := os.Lstat(filePath)
	if err != nil {
		return nil, err
	}
	stat, ok := info.Sys().(*unix.Stat_t)
	if !ok {
		return nil, errors.New("Unable to get the Device ID")
	}
	entry := &FileEntry{
		OrigPath:     filePath,
		Name:         info.Name(),
		Size:         info.Size(),
		IsDir:        info.IsDir(),
		FileMode:     info.Mode(),
		DeletionDate: time.Now(),
		MountRoot:    "",
		DeviceID:     stat.Dev,
		// TrashPath
	}
	// Logic for MountRoot could go here too
	return entry, nil
}

// SetTrashPath is the "second pass" function.
func (f *FileEntry) SetTrashPath(destination string) {
	f.TrashPath = destination
}

func getMountRoot(devId uint64, fileAbsPath string) (string, error) {
	current := fileAbsPath
	for {
		parent := filepath.Dir(current)
		if current == parent {
			return current, nil
		}
		var pStat unix.Stat_t
		if err := unix.Stat(parent, &pStat); err != nil {
			return "", err
		}
		if pStat.Dev != devId {
			return current, nil
		}
		current = parent
	}
}
