package fs

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// "36 35 98:0 / /mnt/data rw,relatime shared:1 - ext4 /dev/sda1 rw,seclabel"

type MountInfo struct {
	MountID      int
	ParentID     int
	Major, Minor int
	Root         string
	MountPoint   string
	Opts         string
	OptFields    string
	FSType       string
	Source       string
	SuperOpts    string
}

func (m *MountInfo) String() string {
	return fmt.Sprintf(
		"%d %d %d:%d %s %s %s %s - %s %s %s",
		m.MountID, m.ParentID, m.Major, m.Minor,
		m.Root, m.MountPoint, m.Opts,
		m.OptFields, m.FSType, m.Source, m.SuperOpts,
	)
}

type FilterFunc func(*MountInfo) (ignore bool)

func ParseMountInfo(r io.Reader, filterFunc FilterFunc) ([]*MountInfo, error) {
	mountinfos := make([]*MountInfo, 0, 20)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		numFields := len(fields)
		if numFields < 10 {
			// should be at least 10 fields
			return nil, fmt.Errorf(
				"parsing '%s' failed: not enough fields (%d)", text, numFields,
			)
		}
		sepIdx := numFields - 4 // "-"
		major, minor, ok := strings.Cut(fields[2], ":")
		if !ok {
			return nil, fmt.Errorf(
				"parsing '%s' failed: unexpected major:minor pair %s",
				text, fields[2],
			)
		}
		info := &MountInfo{
			MountID:    toInt(fields[0]),
			ParentID:   toInt(fields[1]),
			Major:      toInt(major),
			Minor:      toInt(minor),
			Root:       unescape(fields[3]),
			MountPoint: unescape(fields[4]),
			Opts:       fields[5],
			OptFields:  strings.Join(fields[6:sepIdx], " "),
			FSType:     unescape(fields[sepIdx+1]),
			Source:     unescape(fields[sepIdx+2]),
			SuperOpts:  fields[sepIdx+3],
		}

		if filterFunc != nil {
			if ignorable := filterFunc(info); ignorable {
				continue
			}
		}
		mountinfos = append(mountinfos, info)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return mountinfos, nil
}

var (
	tmpFS         = "tmpfs"
	ignoreFsTypes = map[string]bool{
		"proc":        true,
		"sysfs":       true,
		"devtmpfs":    true,
		"configfs":    true,
		"debugfs":     true,
		"tracefs":     true,
		"binfmt_misc": true,
		"fusectl":     true,
		"pstore":      true,
		"devpts":      true,
		"autofs":      true,
		"cgroup":      true,
		"cgroup2":     true,
		"efivarfs":    true,
		"hugetlbfs":   true,
		"mqueue":      true,
		// Network Drives
		"nfs":  true,
		"nfs4": true,
		"cifs": true,
		"smb3": true,
	}
)

func IgnoreFsFunc(minfo *MountInfo) bool {
	if ignoreFsTypes[minfo.FSType] {
		return true
	}
	// Field 6: Mount Options. If 'ro' is present, we can't write a Trash
	// folder.
	if slices.Contains(strings.Split(minfo.Opts, ","), "ro") {
		return true
	}
	return false
}

// toInt converts a string to an int, and ignores any numbers parsing
// errors, as there should not be any.
func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
