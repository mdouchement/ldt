package lualib

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/Shopify/go-lua"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var httpLibrary = []lua.RegistryFunction{
	{
		// http.join("https://localhost", "to", "file")
		Name: "join",
		Function: func(l *lua.State) int {
			var vargs []string
			for i := 1; i <= l.Top(); i++ {
				s, ok := l.ToString(i)
				if !ok {
					lua.Errorf(l, "arg[%d] = %v is not a string", i, l.ToValue(i))
				}
				vargs = append(vargs, s)
			}

			uri, err := url.Parse(vargs[0])
			if err != nil {
				lua.Errorf(l, err.Error())
			}
			uri.Path = path.Join(uri.Path, path.Join(vargs[1:]...))

			l.PushString(uri.String())
			return 1
		},
	},
	{
		// http.download("https://localhost/file", "/tmp/file", true)
		// => true for displaying progress bar
		Name: "download",
		Function: func(l *lua.State) int {
			url := lua.CheckString(l, 1)
			dst := lua.CheckString(l, 2)
			lua.CheckType(l, 3, lua.TypeBoolean)
			showProgress := l.ToBoolean(3)

			resp, err := http.Get(url)
			if err != nil {
				lua.Errorf(l, err.Error())
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				lua.Errorf(l, "bad response status")
			}

			f, err := os.Create(dst)
			if err != nil {
				lua.Errorf(l, err.Error())
			}
			defer f.Close()

			var r io.Reader = resp.Body
			if showProgress {
				defer time.Sleep(500 * time.Millisecond) // just to avoid glitches.
				r = withProgressBar(resp.ContentLength, r)
			}

			if _, err := io.Copy(f, r); err != nil {
				lua.Errorf(l, err.Error())
			}

			if err = f.Sync(); err != nil {
				lua.Errorf(l, err.Error())
			}

			return 0
		},
	},
}

// HTTPOpen opens the http library. Usually passed to Require (local http = require "lualib/http").
func HTTPOpen(l *lua.State) {
	open := func(l *lua.State) int {
		lua.NewLibrary(l, httpLibrary)
		return 1
	}
	lua.Require(l, "lualib/http", open, false)
	l.Pop(1)
}

// -------------
// --------------------------- //
// Utils                       //
// --------------------------- //
// --------

func withProgressBar(size int64, r io.Reader) io.Reader {
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)
	bar := p.AddBar(size,
		mpb.PrependDecorators(
			decor.CountersKibiByte("% 6.1f / % 6.1f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_MMSS, float64(size)/2048),
			decor.Name(" ] "),
			decor.AverageSpeed(decor.UnitKiB, "% .2f"),
		),
	)
	return bar.ProxyReader(r)
}
