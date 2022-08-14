package primitive

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyRF copies a directory.
func CopyRF(src, dst string) (err error) {
	if !IsDir(src) {
		return Copy(src, dst)
	}
	srcs, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	for _, s := range srcs {
		// Get destination path and create directory
		d := strings.Replace(s, src, "", 1)
		if src == "." {
			d = s // Rollback the replacement
		}
		d = filepath.Join(dst, d)
		MkdirAllWithFilename(d)
		//
		err = CopyRF(s, d)
		if err != nil {
			return err
		}
	}
	return
}

// Copy copies a file.
func Copy(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	return out.Sync()
}

// IsDir returns if the given path is a directory or not.
func IsDir(p string) bool {
	s, err := os.Stat(p)
	if err != nil {
		panic(err)
	}
	return s.IsDir()
}

// MkdirAllWithFilename creates a directroy and its parents.
// The filename is exclude from the path.
func MkdirAllWithFilename(p string) {
	MkdirAll(filepath.Dir(p))
}

// MkdirAll creates a directroy and its parents.
func MkdirAll(p string) {
	_ = os.MkdirAll(p, 0755)
}

// Exist returns if a file exists ot not.
func Exist(p string) bool {
	_, err := os.Stat(p)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true // ignoring error
}

// ParseEnviron parses to a map the os.Environ().
func ParseEnviron(environ []string) map[string]string {
	env := make(map[string]string, len(environ))

	for _, kv := range environ {
		kv2 := strings.SplitN(kv, "=", 2)

		key := kv2[0]
		value := kv2[1]

		env[key] = value
	}

	return env
}
