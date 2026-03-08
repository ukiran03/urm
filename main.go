package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// fileExists return it's Absolute path if the file exists, err otherwise
func fileExists(filename string) (string, error) {
	if _, err := os.Stat(filename); err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("Could not determine absolute path: %w", err)
	}
	return absPath, nil
}
