package lualib

import (
	"bufio"
	"strings"

	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago/util"
)

var stringsLibrary = []lua.RegistryFunction{
	{
		// strings.has_prefix("aa:bb", "aa:")
		Name: "has_prefix",
		Function: func(l *lua.State) int {
			s := lua.CheckString(l, 1)
			prefix := lua.CheckString(l, 2)

			l.PushBoolean(strings.HasPrefix(s, prefix))
			return 1
		},
	},
	{
		// strings.has_sufix("aa:bb", ":bb")
		Name: "has_sufix",
		Function: func(l *lua.State) int {
			s := lua.CheckString(l, 1)
			suffix := lua.CheckString(l, 2)

			l.PushBoolean(strings.HasSuffix(s, suffix))
			return 1
		},
	},
	{
		// strings.split("aa:bb", ":")
		Name: "split",
		Function: func(l *lua.State) int {
			s := lua.CheckString(l, 1)
			sep := lua.CheckString(l, 2)

			artifacts := strings.Split(s, sep)
			return util.DeepPush(l, artifacts)
		},
	},
	{
		// strings.join(":", "aa", "bb")
		Name: "join",
		Function: func(l *lua.State) int {
			sep := lua.CheckString(l, 1)

			var args []string
			for i := 2; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				args = append(args, s)
			}

			l.PushString(strings.Join(args, sep))
			return 1
		},
	},
	{
		// strings.lines("line1\nline\r\n")
		Name: "lines",
		Function: func(l *lua.State) int {
			s := lua.CheckString(l, 1)

			var artifacts []string
			scanner := bufio.NewScanner(strings.NewReader(s))
			for scanner.Scan() {
				artifacts = append(artifacts, strings.TrimSpace(scanner.Text()))
			}

			if scanner.Err() != nil {
				lua.Errorf(l, scanner.Err().Error())
			}

			return util.DeepPush(l, artifacts)
		},
	},
}

// StringsOpen opens the Strings library. Usually passed to Require (local ioutil = require "lualib/strings").
func StringsOpen(l *lua.State) {
	open := func(l *lua.State) int {
		lua.NewLibrary(l, stringsLibrary)
		return 1
	}
	lua.Require(l, "lualib/strings", open, false)
	l.Pop(1)
}
