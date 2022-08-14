package lualib

import "github.com/Shopify/go-lua"

// Open opens all lualib libraries.
func Open(l *lua.State) {
	OSOpen(l)
	FilePathOpen(l)
	HTTPOpen(l)
	IOUtilOpen(l)
	YAMLOpen(l)
	StringsOpen(l)
}
