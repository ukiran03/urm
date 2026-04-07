package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type InitAction int

const (
	DoInit InitAction = iota
	NoInit            // Dry run
)

// getHomeTrashPath: returns the valid TrashPath from the current
// user's HomeDir return error otherwise.
func getHomeTrashPath(homePath string) (string, error) {
	homeTrash := filepath.Join(homePath, ".local", "share", "Trash")
	return ensureTrashDir(homeTrash)
}

// getSpecialTrashPath: returns the valid TrashPath for the given
// Mountpoint, either M01 or M02 (creates if needed), return error
// otherwise.
func getSpecialTrashPath(rootPath string, uid int) (string, error) {
	// Method (1): Admin created M01Trash: $topdir/.Trash/$uid
	M01Trash := filepath.Join(rootPath, ".Trash", strconv.Itoa(uid))
	exists, isSymlink, info, err := DirExists(M01Trash)

	// NOTE: MAY also report the user
	if exists && !isSymlink && err == nil {
		sticky, _ := haveSticyBit(info, M01Trash)
		permissioned, _ := havePermissions(info, M01Trash)
		if sticky && permissioned {
			return M01Trash, nil
		}
	}

	// Method (2): Per-User trash M02Trash: $topdir/.Trash-$uid
	M02Trash := filepath.Join(rootPath, fmt.Sprintf(".Trash-%d", uid))
	return ensureTrashDir(M02Trash)
}

// Helper to handle the directory creation logic
func ensureTrashDir(path string) (string, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", &TrashError{Op: "mkdir", Path: path, Err: err}
	}
	return path, nil
}
