package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type SystemTrash struct {
	trash TrashImpl // TrashImpl is an interface
}

func NewSystemTrash(impl TrashImpl) *SystemTrash {
	return &SystemTrash{
		trash: impl,
	}
}

// atomicTrashOperation does few things in atomic fashion like a
// transaction:
//   - makes a file.trashinfo inside $trash/info for given file
//   - moves (renames) the target file to $trash/files
//   - update the $trash/directorysizes
func (syst *SystemTrash) AtomicTrashOperation(targetFile string) (err error) {
	var infoFilePath, dstPath string
	var isInfoFileCreated, isTargetFileMoved bool

	srcPath, err := filepath.Abs(targetFile)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// Rollbacks
			if isInfoFileCreated {
				_ = os.Remove(infoFilePath)
			}
			if isTargetFileMoved {
				// TODO: Try to move it back. If this fails, we have a
				// bigger OS issue.
				_ = os.Rename(dstPath, srcPath)
			}
		}
	}()

	// Action-1: Create info file
	infoFilePath, err = syst.makeInfoFile(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create info file: %w", err)
	}
	isInfoFileCreated = true

	// Action-2: Move target file to trash
	dstPath = strings.TrimSuffix(
		filepath.Base(infoFilePath), filepath.Ext(infoFilePath),
	)
	err = syst.moveTargetToTrash(srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("failed to move file to trash: %w", err)
	}
	isTargetFileMoved = true

	return nil
}

// Parse once at startup for performance
var infoTempl = template.Must(template.New("info").Parse(
	`"[Trash Info]
Path={{ .Path }}
DeletionDate={{ .Time }}"
`),
)

func (syst *SystemTrash) makeInfoFile(absFilePath string) (string, error) {
	trashdir := syst.trash.SourceDir()
	absPath, err := filepath.Abs(absFilePath) // NOTE: Just in case
	if err != nil {
		return "", err
	}

	uniqName := syst.getUniqTrashName(trashdir, filepath.Base(absPath))
	infoPath := filepath.Join(trashdir, "info", uniqName+".trashinfo")

	// Use os.OpenFile with specific permissions (0600 is best practice for trash)
	f, err := os.OpenFile(infoPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Specification uses YYYY-MM-DDThh:mm:ss
	dDate := time.Now().Format("2006-01-02T15:04:05")

	if err := syst.writeInfo(f, absPath, dDate); err != nil {
		return "", err
	}
	return infoPath, nil
}

func (syst *SystemTrash) writeInfo(w io.Writer, origPath, dDate string) error {
	encodedPath := url.PathEscape(origPath) // closest to RFC2396
	data := struct {
		Path string
		Time string
	}{
		Path: encodedPath,
		Time: dDate,
	}
	return infoTempl.Execute(w, data)
}

// [05-03-2026] TODO: this should be a method for each individual trash
// implementation
func (syst *SystemTrash) moveTargetToTrash(absFilePath, newFilename string) error {
	trashdir := syst.trash.SourceDir()
	absPath, err := filepath.Abs(absFilePath) // NOTE: Just in case
	if err != nil {
		return err
	}
	newPathOfTarget := filepath.Join(trashdir, "files", newFilename)
	return os.Rename(absPath, newPathOfTarget)
}

func (syst *SystemTrash) getUniqTrashName(trashdir, filename string) string {
	dst := filepath.Join(trashdir, "files", filename)
	// If it doesn't exist, we are good to go
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return filename
	}
	// If exists, start incrementing
	for i := 2; ; i++ {
		newName := fmt.Sprintf("%s_%d", filename, i) // ex: file.txt_2
		if _, err := os.Stat(newName); os.IsNotExist(err) {
			return newName
		}
	}
}

// https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/4/html-single/introduction_to_system_administration/index#s2-storage-fs-mounting
