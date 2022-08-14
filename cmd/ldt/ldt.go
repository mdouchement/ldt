package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago"
	"github.com/mdouchement/ldt/pkg/lualib"
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

	list bool
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
	return errors.Wrap(lua.DoFile(state, args[0]+".lua"), "could not run action")
}

func listActions() error {
	files, err := filepath.Glob("*.lua")
	if err != nil {
		return errors.Wrap(err, "could not list actions")
	}

	fmt.Println("Available actions")
	fmt.Println("=================")
	for _, file := range files {
		action := file[:len(file)-len(filepath.Ext(file))]
		fmt.Printf("  %s %s\n", appname, action)
	}

	return nil
}
