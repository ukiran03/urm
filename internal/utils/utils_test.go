package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitTrashCan(t *testing.T) {
	// 1. Setup a clean, temporary root for the test
	tmpDir := t.TempDir()
	trashPath := filepath.Join(tmpDir, ".local/share/Trash")

	// 2. Execute the function
	resultPath, err := InitTrashCan(trashPath)
	if err != nil {
		t.Fatalf("InitTrashCan failed: %v", err)
	}

	// 3. Verify the returned path is correct
	if resultPath != trashPath {
		t.Errorf("Expected path %s, got %s", trashPath, resultPath)
	}

	// 4. Define the expected structure to validate
	expectedDirs := []string{
		trashPath,
		filepath.Join(trashPath, "files"),
		filepath.Join(trashPath, "info"),
	}

	for _, path := range expectedDirs {
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Expected directory missing: %s", path)
			continue
		}

		if !info.IsDir() {
			t.Errorf("Expected %s to be a directory, but it is a file", path)
		}

		// 5. Verify Permissions (0700)
		// We mask with 0777 to ignore the setuid/setgid bits
		if mode := info.Mode().Perm(); mode != 0o700 {
			t.Errorf("Path %s has wrong permissions: got %o, want %o",
				path, mode, 0o700)
		}
	}

	// 6. Verify the cache file
	cacheFile := filepath.Join(trashPath, "directorysizes")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Errorf("Cache file %s was not created", cacheFile)
	}
}
