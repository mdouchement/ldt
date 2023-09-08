package lualib

import (
	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago/util"
	"gopkg.in/yaml.v3"
)

var yamlLibrary = []lua.RegistryFunction{
	{
		// yaml.generate({key: "table value"})
		Name: "generate",
		Function: func(l *lua.State) int {
			table, err := util.PullTable(l, 1)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			payload, err := yaml.Marshal(table)
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(string(payload))
			return 1
		},
	},
	{
		// yaml.parse("key: value")
		Name: "parse",
		Function: func(l *lua.State) int {
			payload := lua.CheckString(l, 1)
			var output any
			if err := yaml.Unmarshal([]byte(payload), &output); err != nil {
				lua.Errorf(l, err.Error())
			}

			return util.DeepPush(l, output)
		},
	},
}

// YAMLOpen opens the yaml library. Usually passed to Require (local yaml = require "lualib/yaml").
func YAMLOpen(l *lua.State) {
	open := func(l *lua.State) int {
		lua.NewLibrary(l, yamlLibrary)
		return 1
	}
	lua.Require(l, "lualib/yaml", open, false)
	l.Pop(1)
}
