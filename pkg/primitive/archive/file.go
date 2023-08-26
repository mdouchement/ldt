package archive

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// A File represents the file to archive or to extract.
type File struct {
	fs.FileInfo
	Name       string
	LinkTarget string
	Open       func() (io.ReadCloser, error)
}

// SafeName returns a safe representation of the file's name.
func (f File) SafeName() string {
	sum := blake2b.Sum256([]byte(f.Name))
	return hex.EncodeToString(sum[:])
}

// A FilesFromDiskOptions holds options in order to find files to archive.
type FilesFromDiskOptions struct {
	GlobalPrefix string
	Exclude      *regexp.Regexp
}

// FilesFromDisk returns files archive from the filesystem.
func FilesFromDisk(root string, options FilesFromDiskOptions) ([]*File, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	prefix := options.GlobalPrefix
	if prefix == "" {
		prefix = root
	}

	var files []*File
	var base string

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if options.Exclude != nil && options.Exclude.MatchString(path) {
			return nil
		}

		if d.Type().Type() == fs.ModeSocket {
			return nil
		}

		if d.IsDir() && base == "" {
			base = d.Name() // If root is a folder then it's the first entry of WalkDir
		}

		var file File

		//

		file.FileInfo, err = d.Info()
		if err != nil {
			return err
		}

		//

		file.Name = path
		if options.GlobalPrefix == "" && !d.IsDir() && file.Name == root {
			file.Name = d.Name() // Given only a file as input
		}
		file.Name = strings.TrimPrefix(file.Name, prefix)
		file.Name = filepath.ToSlash(file.Name)
		file.Name = strings.TrimPrefix(file.Name, "/")
		if base != "" {
			file.Name = pathpkg.Join(base, file.Name)
		}

		//

		if d.Type().Type() == os.ModeSymlink {
			// Preserve symlinks
			file.LinkTarget, err = os.Readlink(path)
			if err != nil {
				return fmt.Errorf("%s: readlink: %w", path, err)
			}
		}

		//

		file.Open = func() (io.ReadCloser, error) {
			return os.Open(path)
		}

		//

		files = append(files, &file)
		return nil
	})

	return files, err
}

// FileToDiskHandler is a handler that allows to write files to filesystem.
func FileToDiskHandler(root string) FileHandler {
	return func(t Type, f *File) error {
		target := filepath.Join(root, filepath.FromSlash(f.Name))

		switch t {
		case TypeDirectory:
			if _, err := os.Stat(target); err != nil {
				return os.MkdirAll(target, 0755)
			}

			return nil
		case TypeFile:
			r, err := f.Open()
			if err != nil {
				return err
			}
			defer r.Close()

			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(f.Mode()))
			if err != nil {
				return err
			}
			defer file.Close()

			n, err := io.Copy(file, r)
			if err != nil {
				return err
			}

			if n != f.Size() {
				return fmt.Errorf("%s: bad size (%d->%d)", f.Name, f.Size(), n)
			}

			return file.Sync()
		case TypeSymlink:
			linktarget := filepath.Join(root, filepath.FromSlash(f.LinkTarget))

			if _, err := os.Lstat(linktarget); err == nil {
				if err = os.Remove(linktarget); err != nil {
					return err
				}
			}

			return os.Symlink(target, linktarget)
		case TypeLink:
			linktarget := filepath.Join(root, filepath.FromSlash(f.LinkTarget))

			if _, err := os.Lstat(linktarget); err == nil {
				if err = os.Remove(linktarget); err != nil {
					return err
				}
			}

			return os.Link(target, linktarget)
		}

		return nil
	}
}
