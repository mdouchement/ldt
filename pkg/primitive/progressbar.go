package primitive

import (
	"io"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// WithProgressBar attaches a progress bar to the given io.Reader.
func WithProgressBar(size int64, r io.Reader) io.Reader {
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(50*time.Millisecond),
	)
	bar := p.AddBar(size,
		mpb.PrependDecorators(
			decor.CountersKibiByte("% 6.1f / % 6.1f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_MMSS, float64(size)/2048),
			decor.Name(" ] "),
			decor.AverageSpeed(decor.SizeB1024(0), "% .2f"),
		),
	)
	return bar.ProxyReader(r)
}
