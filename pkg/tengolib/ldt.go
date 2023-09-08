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
	// ldt.halt(msg string)
	"halt": &tengo.UserFunction{
		Name: "halt",
		Value: FuncASR(func(msg string) {
			fmt.Println(msg)
			os.Exit(1)
		}),
	},
	// ldt.catch(...any) => Catcher
	"catch": &tengo.UserFunction{
		Name: "catch",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			var methods *tengo.ImmutableMap
			methods = &tengo.ImmutableMap{
				Value: map[string]tengo.Object{
					// output() => Catcher
					"output": &tengo.UserFunction{
						Name: "output",
						Value: func(nargs ...tengo.Object) (tengo.Object, error) {
							if len(nargs) != 0 {
								return nil, tengo.ErrWrongNumArguments
							}

							for _, arg := range args {
								fmt.Println(arg)
							}

							return methods, nil
						},
					},
					// halt() => undefined
					"halt": &tengo.UserFunction{
						Name: "halt",
						Value: func(nargs ...tengo.Object) (tengo.Object, error) {
							if len(nargs) != 0 {
								return nil, tengo.ErrWrongNumArguments
							}

							for _, arg := range args {
								if _, ok := arg.(*tengo.Error); ok {
									fmt.Println(arg)
									os.Exit(1)
								}
							}

							return tengo.UndefinedValue, nil
						},
					},
					// first() => error/undefined
					"first": &tengo.UserFunction{
						Name: "first",
						Value: func(nargs ...tengo.Object) (tengo.Object, error) {
							if len(nargs) != 0 {
								return nil, tengo.ErrWrongNumArguments
							}

							for _, arg := range args {
								if err, ok := arg.(*tengo.Error); ok {
									return err, nil
								}
							}

							return tengo.UndefinedValue, nil
						},
					},
				},
			}

			return methods, nil
		},
	},
	// ldt.load_direnv(filename string) => error
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
	// ldt.unload_direnv(filename string) => error
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
