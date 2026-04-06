package fsys

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

type MountInfo struct {
	MountID    int    // To uniquely identify the mount
	ParentID   int    // Useful if you need to trace up to the root
	Major      int    // Device Major
	Minor      int    // Device Minor
	MountPoint string // e.g., /mnt/data
	Opts       string // e.g., rw,relatime
	FSType     string // e.g., ext4 (to skip network mounts like nfs/smb)
	Source     string // e.g., /dev/sda1
}

func (m *MountInfo) String() string {
	return fmt.Sprintf(
		"%d %d %d:%d %s %s - %s %s",
		m.MountID, m.ParentID, m.Major, m.Minor,
		m.MountPoint, m.Opts, m.FSType, m.Source,
	)
}

type FilterFunc func(*MountInfo) (ignore bool)

// "36 35 98:0 / /mnt/data rw,relatime shared:1 - ext4 /dev/sda1 rw,seclabel"
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
			MountPoint: unescape(fields[4]),
			Opts:       fields[5],
			FSType:     unescape(fields[sepIdx+1]),
			Source:     unescape(fields[sepIdx+2]),
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
	ignoreFsTypes = map[string]bool{
		// Kernel & System Virtual Filesystems
		"proc":        true, // Process information pseudo-filesystem
		"sysfs":       true, // Kernel objects/subsystems interface
		"devtmpfs":    true, // Device node management
		"configfs":    true, // Userspace-driven kernel object configuration
		"debugfs":     true, // Kernel debugging interface
		"tracefs":     true, // Infrastructure for kernel tracing
		"binfmt_misc": true, // Support for non-native binary formats
		"fusectl":     true, // Control interface for FUSE filesystems
		"pstore":      true, // Persistent storage for kernel oops/logs
		"devpts":      true, // Pseudo-terminal slave devices
		"autofs":      true, // Kernel-based automounter support
		"cgroup":      true, // Control Groups v1
		"cgroup2":     true, // Control Groups v2
		"efivarfs":    true, // Interface to EFI variable storage
		"hugetlbfs":   true, // Huge page memory support
		"mqueue":      true, // POSIX message queues
		"nsfs":        true, // Namespace filesystems
		"rpc_pipefs":  true, // RPC pipe communication for NFS/SunRPC
		"squashfs":    true, // Compressed read-only FS (often used for Snaps)

		// Container & Overlay Filesystems
		"overlay": true, // Docker/Container layered filesystems

		// Network Filesystems
		"nfs":        true, // Network File System (general)
		"nfs4":       true, // Network File System v4
		"cifs":       true, // Common Internet File System (Samba)
		"smb3":       true, // Server Message Block v3
		"davfs":      true, // WebDAV (Remote web folders)
		"fuse.sshfs": true, // Filesystem over SSH

		// Virtual Desktop / FUSE Portals
		"fuse.gvfsd-fuse": true, // GNOME Virtual File System
		"fuse.portal":     true, // Flatpak/Snap document portals
	}

	ignorePrefixes = []string{
		// System Critical Roots
		"/dev",  // Static and dynamic device nodes
		"/proc", // Process and kernel information
		"/sys",  // System hardware and driver information
		"/boot", // Bootloader, kernels, and initrd

		// User Data & Special Exclusions
		"/home",             // The /home mount point itself (prevents .Trash here)
		"/home/docker-data", // Personal exclusion for docker volume data

		// Service & Container Data
		"/var/lib/docker", // Internal Docker storage/images
		"/var/run",        // Modern Linux runtime state (symlink to /run)
		"/run/lock",       // Coordinated device/file locking
		"/run/initramfs",  // Root FS used during early boot

		// Specific User Runtime Portals
		"/run/user/1000/doc",  // Virtual document portals (FUSE)
		"/run/user/1000/gvfs", // Virtual filesystem mounts (FUSE)
	}
)

func IgnoreFsFunc(minfo *MountInfo) bool {
	if ignoreFsTypes[minfo.FSType] {
		return true
	}
	// We want /tmp, but we don't want '/dev/shm' or '/run/user/1000'
	if minfo.FSType == "tmpfs" && minfo.MountPoint != "/tmp" {
		// If it's a RAM disk but NOT the standard /tmp, ignore it.
		// Note: Check your run/media paths; if your Kindle is tmpfs (rare),
		// you might need to adjust this.
		if !strings.HasPrefix(minfo.MountPoint, "/run/media") {
			return true
		}
	}
	// Check the Prefixes
	for _, prefix := range ignorePrefixes {
		if strings.HasPrefix(minfo.MountPoint, prefix) {
			return true
		}
	}
	// Check Read-Only
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
