package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("Error: No input file")
	}
	for _, arg := range args {
		mount, err := GetMountPoint(arg)
		if err != nil {
			log.Printf("%s: %v\n", arg, err)
			continue
		}
		fmt.Printf("%s: %s\n", arg, mount)
	}
}

func GetMountPoint(path string) (string, error) {
	absPath, err := fileCheck(path)
	if err != nil {
		return "", err
	}
	// Get device ID for the current path
	var stat syscall.Stat_t
	if err := syscall.Stat(absPath, &stat); err != nil {
		return "", err
	}
	dev := stat.Dev
	current := absPath
	for {
		parent := filepath.Dir(current)
		var pStat syscall.Stat_t
		if err := syscall.Stat(parent, &pStat); err != nil {
			return "", err
		}
		// If the parent has a different Device ID, 'current' is the
		// mount point
		if pStat.Dev != dev {
			return current, nil
		}
		// Stop if we hit the root directory
		if current == parent {
			return current, nil
		}
		current = parent
	}
}

// fileCheck checks the file existence, and return its absolute path
// if exist otherwise an error.
func fileCheck(file string) (string, error) {
	if _, err := os.Stat(file); err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return "", fmt.Errorf("File is missing: %w", err)
		case errors.Is(err, fs.ErrPermission):
			return "", fmt.Errorf("Permission Denied for %s: %w", file, err)
		default:
			return "", fmt.Errorf("System Error during Stat: %w", err)
		}
	}
	absPath, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("Could Not determine absolute path: %w", err)
	}
	return absPath, nil
}
