package tengolib

import (
	"time"

	"github.com/d5/tengo/v2"
)

var timesModule = map[string]tengo.Object{
	// ldt.duration_format(d int) => string
	"duration_format": &tengo.UserFunction{
		Name: "duration_format",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 1 {
				return nil, tengo.ErrWrongNumArguments
			}

			d, ok := tengo.ToInt64(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}

			duration := time.Duration(d)
			switch {
			case duration > time.Hour:
				duration = duration.Truncate(time.Second)
			case duration > time.Minute:
				duration = duration.Truncate(time.Second)
			case duration > time.Second:
				duration = duration.Truncate(time.Millisecond)
			case duration > time.Millisecond:
				duration = duration.Truncate(time.Microsecond)
			}

			return &tengo.String{Value: duration.String()}, nil
		},
	},
}
