package tengolib

import (
	"os"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
)

func makeOSDirEntry(entry os.DirEntry) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"name": &tengo.UserFunction{
				Name:  "name",
				Value: stdlib.FuncARS(entry.Name),
			},
			"is_dir": &tengo.UserFunction{
				Name:  "is_dir",
				Value: stdlib.FuncARB(entry.IsDir),
			},
			"type": &tengo.UserFunction{
				Name:  "type",
				Value: stdlib.FuncARI64(func() int64 { return int64(entry.Type()) }),
			},
			"info": &tengo.UserFunction{
				Name: "info",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 0 {
						return nil, tengo.ErrWrongNumArguments
					}

					info, err := entry.Info()
					if err != nil {
						return WrapError(err), nil
					}

					return makeOSFileInfo(info), nil
				},
			},
		},
	}
}

func makeOSFileInfo(info os.FileInfo) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"name": &tengo.UserFunction{
				Name:  "name",
				Value: stdlib.FuncARS(info.Name),
			},
			"size": &tengo.UserFunction{
				Name:  "size",
				Value: stdlib.FuncARI64(info.Size),
			},
			"mode": &tengo.UserFunction{
				Name:  "mode",
				Value: stdlib.FuncARI64(func() int64 { return int64(info.Mode()) }),
			},
			"mod_time": &tengo.UserFunction{
				Name: "mod_time",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 0 {
						return nil, tengo.ErrWrongNumArguments
					}
					return &tengo.Time{Value: info.ModTime()}, nil
				},
			},
			"is_dir": &tengo.UserFunction{
				Name:  "is_dir",
				Value: stdlib.FuncARB(info.IsDir),
			},
		},
	}
}
