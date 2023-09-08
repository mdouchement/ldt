package archive

import (
	"archive/tar"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/zeebo/xxh3"
)

// Tar archive constants.
const (
	FormatTar Format = ".tar"

	paxchecksum = "LDT.checksum.xxh3"
)

// A TarReader allows to read a Tar archive.
type TarReader struct {
	r                 *tar.Reader
	checksums         map[string]string
	filenames         map[string]string
	computedChecksums map[string]string
}

// NewTarReader returns a new TarReader.
func NewTarReader(r io.Reader) *TarReader {
	return &TarReader{
		r:                 tar.NewReader(r),
		checksums:         make(map[string]string),
		filenames:         make(map[string]string),
		computedChecksums: make(map[string]string),
	}
}

// Extract reads the archive and calls yield for each entry read.
func (c *TarReader) Extract(yield FileHandler) error {
	for {
		h, err := c.r.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case h == nil:
			continue
		}

		var t Type
		switch h.Typeflag {
		case tar.TypeXGlobalHeader:
			if strings.HasPrefix(h.Name, paxchecksum) {
				for k, v := range h.PAXRecords {
					c.checksums[k] = v
				}
			}
		case tar.TypeDir:
			t = TypeDirectory
		case tar.TypeReg, tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
			t = TypeFile
		case tar.TypeSymlink:
			t = TypeSymlink
		case tar.TypeLink:
			t = TypeLink
		}

		xxh3 := xxh3.New()
		file := &File{
			FileInfo:   h.FileInfo(),
			Name:       h.Name,
			LinkTarget: h.Linkname,
			Open: func() (io.ReadCloser, error) {
				return io.NopCloser(io.TeeReader(c.r, xxh3)), nil
			},
		}

		if err = yield(t, file); err != nil {
			return err
		}

		sname := file.SafeName()
		c.filenames[sname] = file.Name
		c.computedChecksums[sname] = hex.EncodeToString(xxh3.Sum(nil))
	}
}

// Check analyzes the checksums of each entry (must be called after Extract).
func (c *TarReader) Check() error {
	if len(c.checksums) == 0 {
		return nil
	}

	var err error
	for k, v := range c.checksums {
		if c.computedChecksums[k] != v {
			e := fmt.Errorf("corrupted (%s->%s): %s", v, c.computedChecksums[k], c.filenames[k])
			if err == nil {
				err = e
			}
			err = errors.Join(err, e)
		}
	}

	return err
}

//
//
//
//
//

// A TarWriter allows to create a Tar archive.
type TarWriter struct {
	w         *tar.Writer
	checksums map[string]string
}

// NewTarWriter returns a new TarWriter.
func NewTarWriter(w io.Writer) *TarWriter {
	return &TarWriter{
		w:         tar.NewWriter(w),
		checksums: make(map[string]string),
	}
}

// Archives adds files to the archive.
func (c *TarWriter) Archives(files []*File) error {
	for _, file := range files {
		err := c.Archive(file)
		if err != nil {
			return err
		}
	}

	return nil
}

// Archive adds file to the archive.
func (c *TarWriter) Archive(file *File) error {
	//
	// Header
	//

	h, err := tar.FileInfoHeader(file.FileInfo, file.LinkTarget)
	if err != nil {
		return err
	}
	h.Name = file.Name // Complete path

	if err := c.w.WriteHeader(h); err != nil {
		return fmt.Errorf("file %s: writing header: %w", file.Name, err)
	}

	if !slices.Contains([]byte{tar.TypeReg, tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse}, h.Typeflag) {
		// It's not a file.
		return nil
	}

	//
	// Body
	//

	xxh3 := xxh3.New()

	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	w := io.MultiWriter(c.w, xxh3)
	_, err = io.Copy(w, f)

	c.checksums[file.SafeName()] = hex.EncodeToString(xxh3.Sum(nil))
	return err
}

// Close appends checksums to the archive and close the archive.
func (c *TarWriter) Close() error {
	const maxSpecialFileSize = 1<<20 - 128<<10 // Taken from archive/tar package with a safe margin added

	var gpax *tar.Header
	var i, n, idx int
	for k, v := range c.checksums {
		if n == 0 {
			gpax = &tar.Header{
				Typeflag:   tar.TypeXGlobalHeader,
				Name:       fmt.Sprintf("%s.%d", paxchecksum, idx),
				PAXRecords: make(map[string]string),
			}
		}

		gpax.PAXRecords[k] = v

		i++
		n += len(k) + len(v)

		if n >= maxSpecialFileSize || i >= len(c.checksums) {
			if err := c.w.WriteHeader(gpax); err != nil {
				return err
			}
			n = 0
			idx++
		}
	}

	if err := c.w.Flush(); err != nil {
		return err
	}
	return c.w.Close()
}
