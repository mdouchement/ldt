package tengolib

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/d5/tengo/v2"
	"github.com/mdouchement/ldt/pkg/primitive"
)

var httpModule = map[string]tengo.Object{
	// http.join("https://localhost", "to", "file")
	"join": &tengo.UserFunction{
		Name: "join",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) < 2 {
				return nil, tengo.ErrWrongNumArguments
			}

			URL, ok := tengo.ToString(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}

			vargs := make([]string, 0, len(args))
			for idx, arg := range args[1:] {
				p, ok := tengo.ToString(arg)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("args[%d]", idx),
						Expected: "string(compatible)",
						Found:    args[idx].TypeName(),
					}
				}

				vargs = append(vargs, p)
			}

			uri, err := url.Parse(URL)
			if err != nil {
				return wrapError(err), nil
			}
			uri.Path = path.Join(uri.Path, path.Join(vargs...))

			return &tengo.String{Value: uri.String()}, nil
		},
	},
	// http.download("https://localhost/file", "/tmp/file", true)
	// => true for displaying progress bar
	"download": &tengo.UserFunction{
		Name: "download",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 3 {
				return nil, tengo.ErrWrongNumArguments
			}

			url, ok := tengo.ToString(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}

			dst, ok := tengo.ToString(args[1])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "second",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}

			progress, ok := tengo.ToBool(args[2])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "third",
					Expected: "bool(compatible)",
					Found:    args[0].TypeName(),
				}
			}

			//

			resp, err := http.Get(url)
			if err != nil {
				return wrapError(err), nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return wrapError(errors.New("bad response status")), nil
			}

			f, err := os.Create(dst)
			if err != nil {
				return wrapError(err), nil
			}
			defer f.Close()

			var r io.Reader = resp.Body
			if progress {
				defer time.Sleep(500 * time.Millisecond) // just to avoid glitches.
				r = primitive.WithProgressBar(resp.ContentLength, r)
			}

			if _, err := io.Copy(f, r); err != nil {
				return wrapError(err), nil
			}

			if err = f.Sync(); err != nil {
				return wrapError(err), nil
			}

			return tengo.UndefinedValue, nil
		},
	},
}
