package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/parser"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/mdouchement/ldt/pkg/lualib"
	"github.com/mdouchement/ldt/pkg/tengolib"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	appname = "ldt"
)

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"

	list        bool
	extensions  = []string{".tgo", ".tengo", ".lua"} // in precedence order
	mextensions = map[string]func([]string) error{
		".lua":   runlua,
		".tengo": runtengo,
		".tgo":   runtengo,
	}
)

func main() {
	c := &cobra.Command{
		Use:     appname,
		Short:   "Lua dotfiles tool",
		Version: fmt.Sprintf("%s - build %.7s @ %s - %s", version, revision, date, runtime.Version()),
		RunE:    action,
	}
	c.Flags().BoolVarP(&list, "list", "l", false, "List the actions")

	if err := c.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func action(_ *cobra.Command, args []string) error {
	if list || len(args) == 0 {
		return listActions()
	}

	// The extension is already in the action.
	if run, ok := mextensions[filepath.Ext(args[0])]; ok {
		return run(args)
	}

	// The action does not have any extension.
	filenames, err := lookup(args[0])
	if err != nil {
		return errors.Wrap(err, "could not lookup action")
	}
	if len(filenames) == 0 {
		return errors.New("not found")
	}
	args[0] = filenames[0]

	fmt.Println("Using", filenames[0])

	if run, ok := mextensions[filepath.Ext(args[0])]; ok {
		return run(args)
	}

	return errors.New("unsupported action format")
}

func runlua(args []string) error {
	// Initialize Lua's VM and add defaults libraries
	state := lua.NewState()
	lua.OpenLibraries(state)
	goluago.Open(state)
	lualib.Open(state)

	// Forward CLI args to Lua script
	var argv = []string{}
	if len(args) > 1 {
		for _, arg := range args[1:] {
			argv = append(argv, fmt.Sprintf(`"%s"`, arg)) // Surrounding `arg` with double quotes
		}
	}
	stmt := fmt.Sprintf("arg = {%s}", strings.Join(argv, ", ")) // Create `arg` table
	if err := lua.DoString(state, stmt); err != nil {
		return errors.Wrap(err, "could not forward argv")
	}

	// Run the script
	return errors.Wrap(lua.DoFile(state, args[0]), "could not run action")
}

func runtengo(args []string) error {
	// Load modules
	modules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	tengolib.MergeModule(modules, tengolib.AllModuleNames()...)

	// Compile source code
	code, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	fileset := parser.NewFileSet()
	src := fileset.AddFile(filepath.Base(args[0]), -1, len(code))

	p := parser.NewParser(src, code, nil)
	file, err := p.ParseFile()
	if err != nil {
		return err
	}

	c := tengo.NewCompiler(src, nil, nil, modules, nil)
	c.EnableFileImport(true)
	c.SetImportDir(filepath.Dir(args[0]))

	if err := c.Compile(file); err != nil {
		return err
	}

	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()

	// Run the script
	vm := tengo.NewVM(bytecode, nil, -1)
	return vm.Run()
}

func listActions() error {
	filenames, err := lookup("*")
	if err != nil {
		return errors.Wrap(err, "could not list actions")
	}

	dup := make(map[string]int)
	for _, filename := range filenames {
		action := filename[:len(filename)-len(filepath.Ext(filename))]
		dup[action]++
	}

	fmt.Println("Available actions")
	fmt.Println("=================")
	for _, filename := range filenames {
		action := filename[:len(filename)-len(filepath.Ext(filename))]
		if dup[action] > 1 {
			action = filename
		}
		fmt.Printf("  %s %s\n", appname, action)
	}

	return nil
}

func lookup(basename string) ([]string, error) {
	var filenames []string
	for _, ext := range extensions {
		files, err := filepath.Glob(basename + ext)
		if err != nil {
			return nil, err
		}

		filenames = append(filenames, files...)
	}

	return filenames, nil
}
