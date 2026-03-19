package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"ukiran.com/useme/internal/fsys"
)

// getHomeTrashDirectory: returns the HomeTrash of user with setup,
// also creates if absent.
func getHomeTrashDirectory(rootPath string) (string, error) {
	// HomeTrash dir for the user
	var homeTrash string

	trashPath := filepath.Join(rootPath, "Trash")

	_, exists, err := dirExists(trashPath)
	if err != nil {
		return "", fmt.Errorf("failed to check home trash existence: %w", err)
	}

	if exists { // If Trash Dir exists, return it
		return trashPath, nil
	}

	// Otherwise, initialise the Trash Dir
	homeTrash, err = fsys.InitTrashCan(trashPath) // NOTE: Assuming no errors
	if err != nil {
		return "", fmt.Errorf("Error fsys.InitTrashCan in HomeDir: %w", err)
	}
	return homeTrash, nil
}

// getSpecialTrashDirectory: returns the SpecialTrash for the device
// with setup, also creates if absent.
func getSpecialTrashDirectory(rootPath string) (string, error) {
	// SpecialTrash dir for the user ($uid)
	var specialTrash string

	var uidTrashDir string

	// Check for .Trash (Admin created Trash)
	adminTrashDir := filepath.Join(rootPath, ".Trash")

	uidTrashDir, err := getAdminTrashDir(adminTrashDir)
	if err == nil && uidTrashDir != "" {
		return uidTrashDir, nil
	}

	// Create .Trash-1000 ($uid)
	uid := os.Getuid()
	uidTrashDir = filepath.Join(rootPath, fmt.Sprintf(".Trash-%d", uid))

	// If we can't create the directory, we can't use it.
	if err := os.MkdirAll(uidTrashDir, 0o700); err != nil {
		return "",
			fmt.Errorf("failed to create trash directory %s: %w", uidTrashDir, err)
	}

	specialTrash, err = fsys.InitTrashCan(uidTrashDir)
	if err != nil {
		return "", fmt.Errorf("Error fsys.InitTrashCan: %w", err)
	}
	return specialTrash, nil
}

// getAdminTrashDir: checks for a valid ".Trash" dir, which is often
// created by the admin, also should have a sticky-bit and valid
// permissions for the user. Then either return the created/existing
// .Trash/$uid dir, error otherwise.
func getAdminTrashDir(path string) (string, error) {
	info, exists, err := dirExists(path)
	if err != nil || !exists {
		return "", err
	}

	var stickybit, permissioned bool
	stickybit, err = haveSticyBit(path, info)
	if err != nil || !stickybit {
		return "", err
	}
	permissioned, err = havePermissions(path, info)
	if err != nil || !permissioned {
		return "", err
	}

	// create/retrieve the .Trash/$uid dir
	uid := strconv.Itoa(os.Getuid())
	uidDir := filepath.Join(path, uid)

	err = os.MkdirAll(uidDir, 0o700)
	if err != nil {
		return "", fmt.Errorf("failed to create user trash directory: %w", err)
	}
	info, err = os.Stat(uidDir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path %s exists but is not a directory", uidDir)
	}
	return uidDir, nil
}
