package tengolib

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/mdouchement/ldt/pkg/primitive"
	"github.com/mdouchement/ldt/pkg/primitive/archive"
	"github.com/mdouchement/upathex"
)

var osModule = map[string]tengo.Object{
	"chmod_d": &tengo.Int{Value: 0755},
	"chmod_f": &tengo.Int{Value: 0644},
	// os.osname() => string
	"osname": &tengo.UserFunction{
		Name: "osname",
		Value: stdlib.FuncARS(func() string {
			return runtime.GOOS
		}),
	},
	// os.user_exists(username string) => bool
	"user_exists": &tengo.UserFunction{
		Name: "user_exists",
		Value: FuncASRB(func(username string) bool {
			_, err := user.Lookup(username)
			return err == nil
		}),
	},
	// os.user_id(username string) => string/error
	"user_id": &tengo.UserFunction{
		Name: "user_id",
		Value: stdlib.FuncASRSE(func(username string) (string, error) {
			u, err := user.Lookup(username)
			if err != nil {
				return "", err
			}

			return u.Uid, nil
		}),
	},
	// os.group_id(groupname string) => string/error
	"group_id": &tengo.UserFunction{
		Name: "group_id",
		Value: stdlib.FuncASRSE(func(groupname string) (string, error) {
			g, err := user.LookupGroup(groupname)
			if err != nil {
				return "", err
			}

			return g.Gid, nil
		}),
	},
	// os.touch(filename string) => error
	"touch": &tengo.UserFunction{
		Name: "touch",
		Value: stdlib.FuncASRE(func(filename string) error {
			if primitive.Exist(filename) {
				return nil
			}

			f, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer f.Close()

			return nil
		}),
	},
	// os.chown_r(root string, uid string, gid string) => error
	"chown_r": &tengo.UserFunction{
		Name: "chown_r",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 3 {
				return nil, tengo.ErrWrongNumArguments
			}

			root, ok := tengo.ToString(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}
			uid, ok := tengo.ToInt(args[1])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "second",
					Expected: "int(compatible)",
					Found:    args[1].TypeName(),
				}
			}
			gid, ok := tengo.ToInt(args[2])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "third",
					Expected: "int(compatible)",
					Found:    args[2].TypeName(),
				}
			}

			err := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				return os.Chown(path, uid, gid)
			})

			return WrapError(err), nil
		},
	},
	// os.cp(src string, dst string) => error
	"cp": &tengo.UserFunction{
		Name: "cp",
		Value: stdlib.FuncASSRE(func(src, dst string) error {
			return primitive.Copy(src, dst)
		}),
	},
	// os.cp_rf(src string, dst string) => error
	"cp_rf": &tengo.UserFunction{
		Name: "cp_rf",
		Value: stdlib.FuncASSRE(func(src, dst string) error {
			return primitive.CopyRF(src, dst)
		}),
	},
	// os.mv(src string, dst string) => error
	"mv": &tengo.UserFunction{
		Name: "mv",
		Value: stdlib.FuncASSRE(func(src, dst string) error {
			stat, err := os.Stat(dst)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if err == nil && stat.IsDir() {
				dst = filepath.Join(dst, filepath.Base(src))
			}

			return os.Rename(src, dst)
		}),
	},
	// os.write_file(filename string, payload string) => error
	"write_file": &tengo.UserFunction{
		Name: "write_file",
		Value: stdlib.FuncASSRE(func(filename, payload string) error {
			return os.WriteFile(filename, []byte(payload), 0644)
		}),
	},
	// os.checksum(algorithm string, filename string) => string/error
	"checksum": &tengo.UserFunction{
		Name: "checksum",
		Value: FuncASSRSE(func(algorithm, filename string) (string, error) {
			alg := primitive.ChecksumAlg(algorithm)

			f, err := os.Open(filename)
			if err != nil {
				return "", err
			}
			defer f.Close()

			hashes, err := primitive.Checksum(f, alg)
			if err != nil {
				return "", err
			}

			return hex.EncodeToString(hashes[alg].Sum(nil)), nil
		}),
	},
	// os.expand_env(string) => string
	"expand_env": &tengo.UserFunction{
		Name: "expand_env",
		Value: stdlib.FuncASRS(func(path string) string {
			return upathex.ExpandEnv(path)
		}),
	},
	// os.indir(dir string, handler func() error) => error
	// "indir": &tengo.UserFunction{
	// 	Name: "indir",
	// 	Value: func(args ...tengo.Object) (tengo.Object, error) {
	// 		if len(args) < 2 {
	// 			return nil, tengo.ErrWrongNumArguments
	// 		}

	// 		workdir, ok := tengo.ToString(args[0])
	// 		if !ok {
	// 			return nil, tengo.ErrInvalidArgumentType{
	// 				Name:     "first",
	// 				Expected: "string(compatible)",
	// 				Found:    args[0].TypeName(),
	// 			}
	// 		}

	// 		orginal, err := os.Getwd()
	// 		if err != nil {
	// 			return WrapError(err), nil
	// 		}

	// 		if err := os.Chdir(workdir); err != nil {
	// 			return WrapError(err), nil
	// 		}
	// 		defer os.Chdir(orginal) // Should not return an error

	// 		handler, ok := args[1].(*tengo.CompiledFunction)
	// 		if !ok {
	// 			return nil, tengo.ErrInvalidArgumentType{
	// 				Name:     "second",
	// 				Expected: "func(compatible)",
	// 				Found:    args[1].TypeName(),
	// 			}
	// 		}

	// 		return handler.Call() // Await https://github.com/d5/tengo/pull/372
	// 	},
	// },
	// os.archive(name string, path ...string) => error
	"archive": &tengo.UserFunction{
		Name: "archive",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) < 2 {
				return nil, tengo.ErrWrongNumArguments
			}

			name, ok := tengo.ToString(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}

			if !primitive.IsArchiveSupported(name) {
				return WrapError(errors.New("unsupported archive format")), nil
			}

			//

			var filenames []string
			var base string
			for idx, arg := range args[1:] {
				root, ok := tengo.ToString(arg)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("args[%d]", idx),
						Expected: "string(compatible)",
						Found:    args[1+idx].TypeName(),
					}
				}

				info, err := os.Stat(root)
				if err != nil {
					return WrapError(err), nil
				}

				filenames = append(filenames, root)

				if !info.IsDir() {
					root = filepath.Dir(root)
				}

				if base == "" {
					base = root
					continue
				}
				base = primitive.LongestCommonPathPrefix(base, root)
			}

			var files []*archive.File
			for idx, arg := range args[1:] {
				root, ok := tengo.ToString(arg)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("args[%d]", idx),
						Expected: "string(compatible)",
						Found:    args[1+idx].TypeName(),
					}
				}

				fs, err := archive.FilesFromDisk(root, archive.FilesFromDiskOptions{
					GlobalPrefix: base,
					Exclude:      regexp.MustCompile(name + "$"),
				})
				if err != nil {
					return WrapError(err), nil
				}

				files = append(files, fs...)
			}

			//

			f, err := os.Create(name)
			if err != nil {
				return WrapError(err), nil
			}
			defer f.Close()

			codec, err := primitive.NewArchiveWriter(name, f)
			if err != nil {
				return WrapError(err), nil
			}

			if err = codec.Archives(files); err != nil {
				return WrapError(err), nil
			}

			if err = codec.Close(); err != nil {
				return WrapError(err), nil
			}

			if err = f.Sync(); err != nil && !strings.HasSuffix(err.Error(), "operation not supported") {
				return WrapError(err), nil
			}

			return tengo.UndefinedValue, nil
		},
	},
	// os.extract_archive(name string) => error
	"extract_archive": &tengo.UserFunction{
		Name: "extract_archive",
		Value: stdlib.FuncASRE(func(name string) error {
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}

			return extractArchive(name, archive.FileToDiskHandler(pwd))
		}),
	},
	// os.check_archive(name string) => error
	"check_archive": &tengo.UserFunction{
		Name: "check_archive",
		Value: stdlib.FuncASRE(func(name string) error {
			handler := func(t archive.Type, f *archive.File) error {
				if t == archive.TypeDirectory {
					return nil
				}

				r, err := f.Open()
				if err != nil {
					return err
				}
				defer r.Close()

				n, err := io.Copy(io.Discard, r)
				if err != nil {
					return err
				}

				if n != f.Size() {
					return fmt.Errorf("%s: bad size (%d->%d)", f.Name, f.Size(), n)
				}

				return err
			}

			return extractArchive(name, handler)
		}),
	},
}

func extractArchive(name string, handler archive.FileHandler) error {
	if !primitive.IsArchiveSupported(name) {
		return errors.New("unsupported archive format")
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	codec, err := primitive.NewArchiveReader(name, f)
	if err != nil {
		return err
	}

	if err = codec.Extract(handler); err != nil {
		return err
	}

	return codec.Check()
}
