package lualib

import (
	"io/ioutil"

	"github.com/Shopify/go-lua"
)

var ioutilLibrary = []lua.RegistryFunction{
	{
		// ioutil.read_file("go.mod")
		// Deprecated: Use os.read_file instead
		Name: "read_file",
		Function: func(l *lua.State) int {
			data, err := ioutil.ReadFile(lua.CheckString(l, 1))
			if err != nil {
				lua.Errorf(l, err.Error())
			}

			l.PushString(string(data))
			return 1
		},
	},
}

// IOUtilOpen opens the ioutil library. Usually passed to Require (local ioutil = require "lualib/ioutil").
// Deprecated: Use os instead
func IOUtilOpen(l *lua.State) {
	open := func(l *lua.State) int {
		lua.NewLibrary(l, ioutilLibrary)
		return 1
	}
	lua.Require(l, "lualib/ioutil", open, false)
	l.Pop(1)
}
