package tengolib

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/mdouchement/upathex"
)

var filepathModule = map[string]tengo.Object{
	// filepath.dirname("pkg/go.mod")
	"dirname": &tengo.UserFunction{
		Name: "dirname",
		Value: stdlib.FuncASRS(func(path string) string {
			return filepath.Dir(path)
		}),
	},
	// filepath.basename("pkg/go.mod")
	"basename": &tengo.UserFunction{
		Name: "basename",
		Value: stdlib.FuncASRS(func(path string) string {
			return filepath.Base(path)
		}),
	},
	// filepath.join("path", "to", "file")
	"join": &tengo.UserFunction{
		Name: "join",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) < 2 {
				return nil, tengo.ErrWrongNumArguments
			}

			params := make([]string, 0, len(args))
			for idx, arg := range args {
				p, ok := tengo.ToString(arg)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("args[%d]", idx),
						Expected: "string(compatible)",
						Found:    args[idx].TypeName(),
					}
				}

				params = append(params, p)
			}

			return &tengo.String{Value: filepath.Join(params...)}, nil
		},
	},
	// filepath.expand("~/.go/bin/../")
	"expand": &tengo.UserFunction{
		Name: "expand",
		Value: stdlib.FuncASRSE(func(path string) (string, error) {
			// Cleanup separator
			path, err := upathex.ExpandTilde(path)
			if err != nil {
				return "", err
			}

			// Replace environment variables by their values.
			path = upathex.ExpandEnv(path)

			// Compute absolute path.
			return filepath.Abs(path)
		}),
	},
	// filepath.lookup("/home/mdouchement/.go/bin/", ".envrc")
	"lookup": &tengo.UserFunction{
		Name: "lookup",
		Value: FuncASSRSE(func(workdir, filename string) (string, error) {
			var previous string

			for workdir != previous {
				filename := filepath.Join(workdir, filename)

				_, err := os.Stat(filename)
				if err == nil {
					return filename, nil
				}
				if os.IsNotExist(err) {
					previous = workdir
					workdir = filepath.Dir(workdir)
					continue
				}

				return "", err
			}

			return "", errors.New("not found")
		}),
	},
	// filepath.find("~/.go/bin/", ".*image.*", "(?i).*.(jpg|png)$")
	"find": &tengo.UserFunction{
		Name: "find",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) < 2 {
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

			patterns := make([]*regexp.Regexp, 0, len(args)-1)
			for idx, arg := range args[1:] {
				s, ok := tengo.ToString(arg)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("args[%d]", idx),
						Expected: "string(compatible)",
						Found:    args[1+idx].TypeName(),
					}
				}

				r, err := regexp.Compile(s)
				if err != nil {
					return WrapError(err), nil
				}

				patterns = append(patterns, r)
			}

			matches := new(tengo.Array)
			err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				for _, pattern := range patterns {
					if pattern.MatchString(path) {
						matches.Value = append(matches.Value, &tengo.String{Value: path})
						return nil
					}
				}

				return nil
			})
			if err != nil {
				return WrapError(err), nil
			}

			return matches, nil
		},
	},
}
