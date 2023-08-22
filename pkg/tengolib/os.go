package tengolib

import (
	"encoding/hex"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/mdouchement/ldt/pkg/primitive"
	"github.com/mdouchement/upathex"
)

var osModule = map[string]tengo.Object{
	"chmod_d": &tengo.Int{Value: 0755},
	"chmod_f": &tengo.Int{Value: 0644},
	// os.osname()
	"osname": &tengo.UserFunction{
		Name: "osname",
		Value: stdlib.FuncARS(func() string {
			return runtime.GOOS
		}),
	},
	// os.user_exists("myuser")
	"user_exists": &tengo.UserFunction{
		Name: "user_exists",
		Value: FuncASRB(func(username string) bool {
			_, err := user.Lookup(username)
			return err != nil
		}),
	},
	// os.user_id("myuser")
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
	// os.group_id("mygroup")
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
	// os.touch("/tmp/ldt.db")
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
	// os.cp("go.mod", "/tmp")
	"cp": &tengo.UserFunction{
		Name: "cp",
		Value: stdlib.FuncASSRE(func(src, dst string) error {
			return primitive.Copy(src, dst)
		}),
	},
	// os.cp_rf(".", "/tmp/project")
	"cp_rf": &tengo.UserFunction{
		Name: "cp_rf",
		Value: stdlib.FuncASSRE(func(src, dst string) error {
			return primitive.CopyRF(src, dst)
		}),
	},
	// os.mv("/src", "/dst")
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
	// os.write_file("go.mod", payload)
	"write_file": &tengo.UserFunction{
		Name: "write_file",
		Value: stdlib.FuncASSRE(func(filename, payload string) error {
			return os.WriteFile(filename, []byte(payload), 0644)
		}),
	},
	// os.checksum("sha256", "/sbin/nologin")
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
	// os.expand_env("blah blah ${HOME} blah blah")
	"expand_env": &tengo.UserFunction{
		Name: "expand_env",
		Value: stdlib.FuncASRS(func(path string) string {
			return upathex.ExpandEnv(path)
		}),
	},
}
