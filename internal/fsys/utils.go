package fsys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// InitTrashCan: Initialises the TrashDir, with all its requirments
// (files & info directory, dircachefile file)
func InitTrashCan(trashPath string) (string, error) {
	// Ensure the root trash path exists and has strict permissions
	if err := os.MkdirAll(trashPath, 0o700); err != nil {
		return "", fmt.Errorf("failed to create trash root %s: %w", trashPath, err)
	}
	// Explicitly enforce permissions in case the dir already exixted with
	// loose permissions (eg: 0o755)
	if err := os.Chmod(trashPath, 0o700); err != nil {
		return "", fmt.Errorf("failed to secure trash directory: %w", err)
	}

	subdirs := []string{
		filepath.Join(trashPath, "files"), // actual files
		filepath.Join(trashPath, "info"),  // .trashinfo files
	}

	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", fmt.Errorf(
				"failed to create trash subdir %s: %w", dir, err,
			)
		}
	}
	dirCacheFile := filepath.Join(trashPath, "directorysizes")
	if err := MakeDirCacheFile(dirCacheFile); err != nil {
		return "", fmt.Errorf("Failed to create directorysizes file: %w", err)
	}
	return trashPath, nil
}

func MakeDirCacheFile(cachepath string) error {
	// O_CREATE | O_EXCL ensures we don't truncate or touch it if it
	// already exists
	f, err := os.OpenFile(cachepath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return fmt.Errorf("could not create cache file %s: %w", cachepath, err)
	}
	return f.Close()
}
