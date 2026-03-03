package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type TrashImpl interface {
	TrashDir() (string, error)
}

type HomeTrash struct {
	homeDir  string
	trashDir string
}

func NewHomeTrash() (*HomeTrash, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	return &HomeTrash{homeDir: home}, nil
}

func (ht *HomeTrash) TrashDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(ht.homeDir, ".local", "share")
	}
	trashPath := filepath.Join(dataHome, "Trash")
	return setupTrashDir(trashPath)
}

func setupTrashDir(path string) (string, error) {
	// Ensure the root trash path exists and has strict permissions
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", fmt.Errorf("failed to create trash root %s: %w", path, err)
	}
	// Explicitly enforce permissions in case the dir already exixted with
	// loose permissions (eg: 0o755)
	if err := os.Chmod(path, 0o700); err != nil {
		return "", fmt.Errorf("failed to secure trash directory: %w", err)
	}

	subdirs := []string{
		filepath.Join(path, "files"), // actual files
		filepath.Join(path, "info"),  // metadata files
	}
	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", fmt.Errorf(
				"failed to create trash subdir %s: %w", dir, err,
			)
		}
	}
	dirCacheFile := filepath.Join(path, "directorysizes")
	if err := makeDirCacheFile(dirCacheFile); err != nil {
		return "", err
	}
	return path, nil
}

func makeDirCacheFile(cachepath string) error {
	// O_CREATE | O_EXCL ensures we don't truncate or touch it if it exists
	f, err := os.OpenFile(cachepath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return fmt.Errorf("could not create cache file %s: %w", cachepath, err)
	}
	return f.Close()
}

type SpecialTrash struct {
	trashDir string
}

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		// Path exists, let's verify it's actually a directory
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		// Path explicitly does not exist
		return false, nil
	}
	// An error occurred (e.g., permission issues)
	return false, err
}
