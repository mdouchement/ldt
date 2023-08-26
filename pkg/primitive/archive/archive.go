package archive

// A Format is the ext name of a archive.
type Format string

// A Type defines a datatype archived.
type Type string

// All supported types.
const (
	TypeDirectory Type = "directory"
	TypeFile      Type = "file"
	TypeSymlink   Type = "symlink"
	TypeLink      Type = "link"
)

// A FileHandler is called when an archive is read.
type FileHandler func(Type, *File) error

// A Reader is used to read archive files.
type Reader interface {
	Extract(FileHandler) error
	Check() error
}

// An Writer is used to create archive files.
type Writer interface {
	Archives([]*File) error
	Archive(*File) error
	Close() error
}
