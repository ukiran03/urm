package fsys

import (
	"fmt"
	"os"
	"strings"
)

// Source: github.com/moby/sys/mountinfo/mountinfo_linux.go
//
// This function converts all such escape sequences back to ASCII, and
// returns the unescaped string.
func unescape(path string) string {
	// Try to avoid copying.
	if strings.IndexByte(path, '\\') == -1 {
		return path
	}

	// The following code is UTF-8 transparent as it only looks for
	// some specific characters (backslash and 0..7) with values less
	// than utf8.RuneSelf, and everything else is passed through as is.
	buf := make([]byte, len(path))
	bufLen := 0
	for i := 0; i < len(path); i++ {
		c := path[i]
		// Look for \NNN, i.e. a backslash followed by three octal
		// digits. Maximum value is 177 (equals utf8.RuneSelf-1).
		if c == '\\' && i+3 < len(path) &&
			(path[i+1] == '0' || path[i+1] == '1') &&
			(path[i+2] >= '0' && path[i+2] <= '7') &&
			(path[i+3] >= '0' && path[i+3] <= '7') {
			// Convert from ASCII to numeric values.
			c1 := path[i+1] - '0'
			c2 := path[i+2] - '0'
			c3 := path[i+3] - '0'
			// Each octal digit is three bits, thus the shift value.
			c = c1<<6 | c2<<3 | c3
			// We read three extra bytes of input.
			i += 3
		}
		buf[bufLen] = c
		bufLen++
	}

	return string(buf[:bufLen])
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
