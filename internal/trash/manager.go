package trash

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"ukiran.com/useme/internal/fs"
)

type TrashManager struct {
	HomeTrash *TrashCan
	Mounts    map[string]*TrashCan // Keyed by mount point path
}

var (
	mountinfoFile  = "/proc/self/mountinfo"
	userHomeDir, _ = os.UserHomeDir()
)

func getMountsInfos() ([]*fs.MountInfo, error) {
	f, err := os.Open(mountinfoFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening mountinfo file: %w", err)
	}
	mounts, err := fs.ParseMountInfo(f, fs.IgnoreFsFunc)
	if err != nil {
		return nil, fmt.Errorf("Error parsing mountinfo file: %w", err)
	}
	return mounts, nil
}

func NewTrashManager(mounts []*fs.MountInfo, homeDir string) (*TrashManager, error) {
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
	// HomeTrash is a special case; it's not a partition root,
	// it's a specific folder.
	tm.HomeTrash = &TrashCan{
		RootPath: dataHome,
	}
	return tm, nil
}

// FindTarget decides which TrashCan should be used.
func (tm *TrashManager) FindTarget(filent *FileEntry) *TrashCan {
	// Get device ID of filePath
	// If same as Home, return tm.HomeTrash
	// Else, lookfor/initialize a TrashCan on that mount point

	if filent.DeviceID == tm.HomeTrash.DeviceID {
		return tm.HomeTrash
	} else {
		if trash, ok := tm.Mounts[filent.MountRoot]; ok {
			return trash
		}
	}
	// TODO: start from here
	return &TrashCan{}
}

type MoveStrategy int

const (
	MoveAtomic   MoveStrategy = iota // Standard os.Rename
	MoveFallback                     // Copy to Home + Delete source
	MoveIdentify                     // Just identify, don't move (dry run)
)

func (tm *TrashManager) Put(filent *FileEntry, strategy MoveStrategy) error {
	targetCan := tm.FindTarget(filent)

	if onSameDevice(filent.MountRoot, targetCan) {
		// return targetCan.Put(filePath) // TODO:
	}

	switch strategy {
	case MoveFallback:
		// return tm.CopyAndDelete(filePath, tm.HomeTrash) // TODO:
	default:
		return errors.New("cross-device link: use Fallback strategy")
	}
	return nil
}

// func onSameDevice(dstDevId, srcDevId uint64) bool
func onSameDevice(fpath string, trash *TrashCan) bool
