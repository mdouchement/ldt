package lualib

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/Shopify/go-lua"
	"github.com/direnv/direnv/v2/pkg/dotenv"
	"github.com/mdouchement/ldt/pkg/primitive"
	"github.com/mdouchement/upathex"
)

var osLibrary = []lua.RegistryFunction{
	{
		// os.osname()
		Name: "osname",
		Function: func(l *lua.State) int {
			l.PushString(runtime.GOOS)
			return 1
		},
	},
	{
		// os.user_exists("myuser")
		Name: "user_exists",
		Function: func(l *lua.State) int {
			username := lua.CheckString(l, 1)
			_, err := user.Lookup(username)

			l.PushBoolean(err == nil)
			return 1
		},
	},
	{
		// os.user_id("myuser")
		Name: "user_id",
		Function: func(l *lua.State) int {
			username := lua.CheckString(l, 1)
			u, err := user.Lookup(username)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(u.Uid)
			return 1
		},
	},
	{
		// os.group_id("mygroup")
		Name: "group_id",
		Function: func(l *lua.State) int {
			groupname := lua.CheckString(l, 1)
			g, err := user.LookupGroup(groupname)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(g.Gid)
			return 1
		},
	},
	{
		// os.touch("/tmp/ldt.db")
		Name: "touch",
		Function: func(l *lua.State) int {
			filename := lua.CheckString(l, 1)

			if primitive.Exist(filename) {
				return 0
			}

			f, err := os.Create(filename)
			if err != nil {
				lua.Errorf(l, err.Error())
			}
			defer f.Close()

			return 0
		},
	},
	{
		// os.exist("~/tmp/binary")
		Name: "exist",
		Function: func(l *lua.State) int {
			path := lua.CheckString(l, 1)
			l.PushBoolean(primitive.Exist(path))

			return 1
		},
	},
	{
		// os.chmod("~/tmp/binary", 0755)
		Name: "chmod",
		Function: func(l *lua.State) int {
			name := lua.CheckString(l, 1)
			mode, err := strconv.ParseUint(strconv.Itoa(lua.CheckInteger(l, 2)), 8, 32) // convert base10 to base8
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			if err = os.Chmod(name, os.FileMode(mode)); err != nil {
				lua.Errorf(l, err.Error())
			}
			return 0
		},
	},
	{
		// os.chown("~/tmp/binary", 1001, 1001, true)  -> recursive
		// os.chown("~/tmp/binary", 1001, 1001, false) -> not recursive
		Name: "chown",
		Function: func(l *lua.State) int {
			root := lua.CheckString(l, 1)
			uid := lua.CheckInteger(l, 2)
			gid := lua.CheckInteger(l, 3)
			recursive := l.ToBoolean(4)

			if !recursive {
				if err := os.Chown(root, uid, gid); err != nil {
					lua.Errorf(l, err.Error())
				}

				return 0
			}

			err := filepath.Walk(root, func(path string, _ os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				return os.Chown(path, uid, gid)
			})

			if err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.symlink("/tmp/ldt.db", "link.db")
		Name: "symlink",
		Function: func(l *lua.State) int {
			src := lua.CheckString(l, 1)
			dst := lua.CheckString(l, 2)
			if err := os.Symlink(src, dst); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.mkdir("~/tmp/something")
		Name: "mkdir",
		Function: func(l *lua.State) int {
			folder := lua.CheckString(l, 1)
			if err := os.Mkdir(folder, 0755); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.mkdir_p("~/tmp/something")
		Name: "mkdir_p",
		Function: func(l *lua.State) int {
			folder := lua.CheckString(l, 1)
			if err := os.MkdirAll(folder, 0755); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.cp("go.mod", "/tmp")
		Name: "cp",
		Function: func(l *lua.State) int {
			src := lua.CheckString(l, 1)
			dst := lua.CheckString(l, 2)
			if err := primitive.Copy(src, dst); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.cp_rf(".", "/tmp/project")
		Name: "cp_rf",
		Function: func(l *lua.State) int {
			src := lua.CheckString(l, 1)
			dst := lua.CheckString(l, 2)
			if err := primitive.CopyRF(src, dst); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.mv("/src", "/dst")
		Name: "mv",
		Function: func(l *lua.State) int {
			src := lua.CheckString(l, 1)
			dst := lua.CheckString(l, 2)

			if primitive.Exist(dst) {
				stat, err := os.Stat(dst)
				if err != nil {
					lua.Errorf(l, err.Error())
				}
				if stat.IsDir() {
					dst = filepath.Join(dst, filepath.Base(src))
				}
			}

			if err := os.Rename(src, dst); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.cp("go.mod")
		Name: "rm",
		Function: func(l *lua.State) int {
			filename := lua.CheckString(l, 1)
			if err := os.Remove(filename); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.cp("/tmp/project")
		Name: "rm_rf",
		Function: func(l *lua.State) int {
			dst := lua.CheckString(l, 1)
			if err := os.RemoveAll(dst); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.exec("useradd", "--no-create-home", "--shell", "/sbin/nologin", "myuser")
		Name: "exec",
		Function: func(l *lua.State) int {
			name := lua.CheckString(l, 1)

			var args []string
			for i := 2; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				args = append(args, s)
			}

			cmd := exec.Command(name, args...)
			std, err := cmd.CombinedOutput()

			if err != nil {
				lua.Errorf(l, err.Error())
			}

			if len(std) > 0 {
				fmt.Println(string(std))
			}

			l.PushString(string(std))
			return 1
		},
	},
	{
		// os.exec_in("/workdir", "useradd", "--no-create-home", "--shell", "/sbin/nologin", "myuser")
		Name: "exec_in",
		Function: func(l *lua.State) int {
			workdir := lua.CheckString(l, 1)
			name := lua.CheckString(l, 2)

			var args []string
			for i := 3; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				args = append(args, s)
			}

			cmd := exec.Command(name, args...)
			cmd.Dir = workdir
			std, err := cmd.CombinedOutput()
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			if len(std) > 0 {
				fmt.Println(string(std))
			}

			l.PushString(string(std))
			return 1
		},
	},
	{
		// local stdout, stderr = os.exec_catched("useradd", "--no-create-home", "--shell", "/sbin/nologin", "myuser")
		Name: "exec_catched",
		Function: func(l *lua.State) int {
			name := lua.CheckString(l, 1)

			var args []string
			for i := 2; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				args = append(args, s)
			}

			cmd := exec.Command(name, args...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err := cmd.Run()

			//

			if stdout.Len() != 0 {
				l.PushString(stdout.String())
			} else {
				l.PushNil()
			}

			//

			if err != nil {
				std := stderr.String()
				stderr.Reset()
				stderr.WriteString(err.Error() + ": " + std)
			}

			if stderr.Len() != 0 {
				l.PushString(stderr.String())
			} else {
				l.PushNil()
			}

			//

			return 2
		},
	},
	{
		// os.read_file("go.mod")
		Name: "read_file",
		Function: func(l *lua.State) int {
			payload, err := os.ReadFile(lua.CheckString(l, 1))
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(string(payload))
			return 1
		},
	},
	{
		// os.write_file("go.mod", payload)
		Name: "write_file",
		Function: func(l *lua.State) int {
			filename := lua.CheckString(l, 1)
			payload := lua.CheckString(l, 2)

			err := os.WriteFile(filename, []byte(payload), 0644)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
	{
		// os.checksum("sha256", "/sbin/nologin")
		Name: "checksum",
		Function: func(l *lua.State) int {
			alg := primitive.ChecksumAlg(lua.CheckString(l, 1))
			filename := lua.CheckString(l, 2)

			f, err := os.Open(filename)
			if err != nil {
				lua.Errorf(l, err.Error())
			}
			defer f.Close()

			hashes, err := primitive.Checksum(f, alg)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(hex.EncodeToString(hashes[alg].Sum(nil)))
			return 1
		},
	},
	{
		// os.expand_env("blah blah ${HOME} blah blah")
		Name: "expand_env",
		Function: func(l *lua.State) int {
			path := lua.CheckString(l, 1)

			path = upathex.ExpandEnv(path)

			l.PushString(path)
			return 1
		},
	},
	{
		// os.load_direnv(filename)
		Name: "load_direnv",
		Function: func(l *lua.State) int {
			filename := lua.CheckString(l, 1)

			if _, ok := envcache[filename]; ok {
				lua.Errorf(l, "%s already in use", filename)
			}

			data, err := os.ReadFile(filename)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			envmap, err := dotenv.Parse(string(data))
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			env := primitive.NewEnv(envmap)
			env.Export()

			envcache[filename] = env
			return 0
		},
	},
	{
		// os.unload_direnv(filename)
		Name: "unload_direnv",
		Function: func(l *lua.State) int {
			filename := lua.CheckString(l, 1)

			env, ok := envcache[filename]
			if !ok {
				lua.Errorf(l, "%s not loaded", filename)
			}

			env.Restore()
			delete(envcache, filename)

			return 0
		},
	},
}

var envcache map[string]*primitive.Env

// OSOpen opens the os library. Usually passed to Require (local os = require "lualib/os").
func OSOpen(l *lua.State) {
	envcache = make(map[string]*primitive.Env)

	open := func(l *lua.State) int {
		lua.NewLibrary(l, osLibrary)
		return 1
	}
	lua.Require(l, "lualib/os", open, false)
	l.Pop(1)
}
