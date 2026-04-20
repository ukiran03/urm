package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

// InitTrashCan: Initialises the TrashDir, with all its requirments
// (files, info directories and dircachefile file)
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
	if err := makeDirCacheFile(dirCacheFile); err != nil {
		return "", fmt.Errorf("Failed to create directorysizes file: %w", err)
	}
	return trashPath, nil
}

func makeDirCacheFile(cachepath string) error {
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

// TODO: This is not correct/safe impl, needs refactoring
func ConcurrnetDirSize(path string) (int64, error) {
	var total atomic.Int64
	var wg sync.WaitGroup
	sema := make(chan struct{}, 2*runtime.NumCPU())

	var walker func(string) error
	walker = func(p string) error {
		sema <- struct{}{}        // acquire
		defer func() { <-sema }() // defer release

		entries, err := os.ReadDir(p)
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				fmt.Fprintf(os.Stderr, "Permission denied, skipping [%s]\n", p)
				return nil // continue with siblings
			}
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", p, err)
			return err
		}

		for _, entry := range entries {
			fullpath := filepath.Join(p, entry.Name())
			if entry.IsDir() {
				wg.Go(func() {
					// ignore error for now (or propagate via channel)
					_ = walker(fullpath)
				})
			} else {
				info, err := entry.Info()
				if err == nil {
					total.Add(info.Size())
				}
			}
		}
		return nil
	}

	wg.Go(func() {
		walker(path)
	})
	wg.Wait()
	return total.Load(), nil
}

func FileExists(filename string) (string, bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return "", false, fmt.Errorf("file %s does not exist", filename)
		case os.IsPermission(err):
			return "", false,
				fmt.Errorf("you don't have persmission to access %s", filename)
		default:
			return "", false, err
		}
	}
	defer f.Close()
	return f.Name(), true, nil
}
