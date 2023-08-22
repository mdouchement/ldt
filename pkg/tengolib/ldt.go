package tengolib

import (
	"fmt"
	"os"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/direnv/direnv/v2/pkg/dotenv"
	"github.com/mdouchement/ldt/pkg/primitive"
)

var envcache = make(map[string]*primitive.Env)

var ldtModule = map[string]tengo.Object{
	// ldt.halt("halt message")
	"halt": &tengo.UserFunction{
		Name: "halt",
		Value: FuncASR(func(msg string) {
			fmt.Println(msg)
			os.Exit(1)
		}),
	},
	// ldt.load_direnv(filename)
	"load_direnv": &tengo.UserFunction{
		Name: "load_direnv",
		Value: stdlib.FuncASRE(func(filename string) error {
			if _, ok := envcache[filename]; ok {
				return fmt.Errorf("%s already in use", filename)
			}

			data, err := os.ReadFile(filename)
			if err != nil {
				return err
			}

			envmap, err := dotenv.Parse(string(data))
			if err != nil {
				return err
			}

			env := primitive.NewEnv(envmap)
			env.Export()

			envcache[filename] = env

			return nil
		}),
	},
	// ldt.unload_direnv(filename)
	"unload_direnv": &tengo.UserFunction{
		Name: "unload_direnv",
		Value: stdlib.FuncASRE(func(filename string) error {
			env, ok := envcache[filename]
			if !ok {
				return fmt.Errorf("%s not loaded", filename)
			}

			env.Restore()
			delete(envcache, filename)
			return nil
		}),
	},
}
