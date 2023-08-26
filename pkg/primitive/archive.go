package primitive

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/mdouchement/ldt/pkg/primitive/archive"
)

func NewArchiveReader(name string, r io.Reader) (archive.Reader, error) {
	switch archive.Format(filepath.Ext(name)) {
	case archive.FormatTar:
		return archive.NewTarReader(r), nil
	default:
		return nil, errors.New("unsupported archive format")
	}
}

func NewArchiveWriter(name string, w io.Writer) (archive.Writer, error) {
	switch archive.Format(filepath.Ext(name)) {
	case archive.FormatTar:
		return archive.NewTarWriter(w), nil
	default:
		return nil, errors.New("unsupported archive format")
	}
}

func IsArchiveSupported(name string) bool {
	switch archive.Format(filepath.Ext(name)) {
	case archive.FormatTar:
		return true
	default:
		return false
	}
}
