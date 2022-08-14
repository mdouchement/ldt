package lualib

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago/util"
	"github.com/mdouchement/upathex"
)

var filepathLibrary = []lua.RegistryFunction{
	{
		// filepath.dirname("pkg/go.mod")
		Name: "dirname",
		Function: func(l *lua.State) int {
			path := lua.CheckString(l, 1)
			l.PushString(filepath.Dir(path))
			return 1
		},
	},
	{
		// filepath.basename("pkg/go.mod")
		Name: "basename",
		Function: func(l *lua.State) int {
			path := lua.CheckString(l, 1)
			l.PushString(filepath.Base(path))
			return 1
		},
	},
	{
		// filepath.join("path", "to", "file")
		Name: "join",
		Function: func(l *lua.State) int {
			var vargs []string
			for i := 1; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				vargs = append(vargs, s)
			}

			l.PushString(filepath.Join(vargs...))
			return 1
		},
	},
	{
		// filepath.expand("~/.go/bin/../")
		Name: "expand",
		Function: func(l *lua.State) int {
			path := lua.CheckString(l, 1)

			// Cleanup separator
			path, err := upathex.ExpandTilde(path)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			// Replace environment variables by their values.
			path = upathex.ExpandEnv(path)

			// Compute absolute path.
			path, err = filepath.Abs(path)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(path)
			return 1
		},
	},
	{
		// filepath.expand_tilde("~/.go/bin/")
		Name: "expand_tilde",
		Function: func(l *lua.State) int {
			path := lua.CheckString(l, 1)

			path, err := upathex.ExpandTilde(path)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(path)
			return 1
		},
	},
	{
		// filepath.lookup("/home/mdouchement/.go/bin/", ".envrc")
		Name: "lookup",
		Function: func(l *lua.State) int {
			workdir := lua.CheckString(l, 1)
			filename := lua.CheckString(l, 2)

			var previous string

			for workdir != previous {
				filename := filepath.Join(workdir, filename)

				_, err := os.Stat(filename)
				if err == nil {
					l.PushString(filename)
					return 1
				}
				if os.IsNotExist(err) {
					previous = workdir
					workdir = filepath.Dir(workdir)
					continue
				}

				lua.Errorf(l, err.Error())
			}

			lua.Errorf(l, "not found")
			return 0
		},
	},
	{
		// filepath.find("~/.go/bin/", ".*image.*", "(?i).*.(jpg|png)$")
		Name: "find",
		Function: func(l *lua.State) int {
			root := lua.CheckString(l, 1)

			var patterns []*regexp.Regexp
			for i := 2; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				r, err := regexp.Compile(s)
				if err != nil {
					lua.Errorf(l, "pattern: %s => %s", s, err.Error())
				}
				patterns = append(patterns, r)
			}

			var matches []string
			err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				for _, pattern := range patterns {
					if pattern.MatchString(path) {
						matches = append(matches, path)
						return nil
					}
				}

				return nil
			})
			if err != nil {
				lua.Errorf(l, "walk: %s", err.Error())
			}

			return util.DeepPush(l, matches)
		},
	},
}

// FilePathOpen opens the filepath library. Usually passed to Require (local filepath = require "lualib/filepath").
func FilePathOpen(l *lua.State) {
	open := func(l *lua.State) int {
		lua.NewLibrary(l, filepathLibrary)
		return 1
	}
	lua.Require(l, "lualib/filepath", open, false)
	l.Pop(1)
}
