package trash

import (
	"fmt"
	"os"
	"path/filepath"

	"ukiran.com/useme/internal/fsys"
)

type TrashManager struct {
	HomeTrash *TrashCan
	Mounts    map[string]*TrashCan // Keyed by mount point path
}

var (
	mountinfoFile  = "/proc/self/mountinfo"
	userHomeDir, _ = os.UserHomeDir()
)

func NewTrashManager(mounts []*fsys.MountInfo, homeDir string) *TrashManager {
	tm := &TrashManager{
		Mounts: make(map[string]*TrashCan, len(mounts)),
	}
	for _, m := range mounts {
		tm.Mounts[m.MountPoint] = NewTrashCan(m)
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	tm.HomeTrash = &TrashCan{
		RootPath: dataHome,
	}
	return tm
}

// FindTarget decides which TrashCan should be used.
// Get device ID of filePath If same as Home, return tm.HomeTrash
// Else, look for a TrashCan on that mount point
func (tm *TrashManager) FindTarget(filent *FileEntry) (*TrashCan, error) {
	if filent.DeviceID == tm.HomeTrash.DeviceID {
		return tm.HomeTrash, nil
	} else {
		if trash, ok := tm.Mounts[filent.MountRoot]; ok &&
			onSameDevice(filent.DeviceID, trash.DeviceID) { // Is this neccessary?
			return trash, nil
		}
	}
	return nil, fmt.Errorf("No trashCan found for: %s", filent.Name)
}

type MoveStrategy int

const (
	MoveAtomic   MoveStrategy = iota // Standard os.Rename
	MoveFallback                     // Copy to Home + Delete source
	MoveIdentify                     // Just identify, don't move (dry run)
)

func (tm *TrashManager) Put(filent *FileEntry, strategy MoveStrategy) error {
	targetCan, err := tm.FindTarget(filent)
	if err != nil {
		switch strategy {
		case MoveFallback:
			// Prompt to user and operate
			return tm.CopyAndDelete(filent)
		}
	}
	return targetCan.Move(filent)
}

func (tm *TrashManager) CopyAndDelete(filent *FileEntry) error {
	// TODO: Copy to HomeTrash and delete the source file (filent)
	panic("unimplemented")
}
