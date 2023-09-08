package tengolib

import "github.com/d5/tengo/v2"

// FuncASR transforms a function of 'func(string)' signature into
// CallableFunc type. User function will return 'true' if underlying native
// function returns nil.
func FuncASR(fn func(string)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}

		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		fn(s1)
		return tengo.UndefinedValue, nil
	}
}

// FuncASSRSE transforms a function of 'func(string, string) (string, error)' signature into
// CallableFunc type. User function will return 'true' if underlying native
// function returns nil.
func FuncASSRSE(fn func(string, string) (string, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}

		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s, err := fn(s1, s2)
		if err != nil {
			return wrapError(err), nil
		}

		return &tengo.String{Value: s}, nil
	}
}

func wrapError(err error) tengo.Object {
	if err == nil {
		return tengo.TrueValue
	}
	return &tengo.Error{Value: &tengo.String{Value: err.Error()}}
}

// FuncASRB transforms a function of 'func(string) bool' signature into
// CallableFunc type. User function will return 'true' if underlying native
// function returns nil.
func FuncASRB(fn func(string) bool) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}

		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		if fn(s1) {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
}
