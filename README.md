# Lua Dotfiles Tool

A dead simple tool that allows to run Lua's scripts. It provides several libraries.

It can run lua scripts (with compatible `require`) for doing whatever you want.

## Usage

```
$ cat install-trololo.lua
local sysos = require "os"
local os = require "lualib/os"
local filepath = require "lualib/filepath"
local ioutil = require "lualib/ioutil"
local yaml = require "lualib/yaml"
local http = require "lualib/http"

if #arg ~= 1 then
    print("You must provide your `dotfiles.yml' as the only argument")
    sysos.exit(1)
end

local config_path = filepath.expand(arg[1])
local config = yaml.parse(ioutil.read_file(config_path))

for section, entries in pairs(config) do
    print("== "..section.." ==")
end
```
```
$ ldt install-trololo myfile.yml
```

## Libraries

- [core](https://github.com/Shopify/go-lua) provided by Shopify go-lua project.
- [goluago](https://github.com/Shopify/goluago)
- `pkg/lualib` that provide bunch of helpers