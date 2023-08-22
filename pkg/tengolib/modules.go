package tengolib

import (
	"github.com/d5/tengo/v2"
)

// BuiltinModules are builtin type standard library modules.
var BuiltinModules = map[string]map[string]tengo.Object{
	"filepath": filepathModule,
	"http":     httpModule,
	"ldt":      ldtModule,
	"os":       osModule, // Missing functions from github.com/d5/tengo/v2/stdlib
	"yaml":     yamlModule,
}

// AllModuleNames returns a list of all default module names.
func AllModuleNames() []string {
	var names []string
	for name := range BuiltinModules {
		names = append(names, name)
	}
	return names
}

// GetModuleMap returns the module map that includes all modules
// for the given module names.
func GetModuleMap(names ...string) *tengo.ModuleMap {
	modules := tengo.NewModuleMap()
	for _, name := range names {
		if mod := BuiltinModules[name]; mod != nil {
			modules.AddBuiltinModule(name, mod)
		}
	}
	return modules
}

// MergeModule appends or merges to the given modules all the given module names.
func MergeModule(modules *tengo.ModuleMap, names ...string) {
	for _, name := range names {
		if mod := BuiltinModules[name]; mod != nil {
			// Merge with existing one
			m := modules.GetBuiltinModule(name)
			if m != nil {
				for k, v := range osModule {
					m.Attrs[k] = v
				}

				continue
			}

			// Add the new module
			modules.AddBuiltinModule(name, mod)
		}
	}
}
