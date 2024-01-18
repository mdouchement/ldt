package tengolib

import (
	"fmt"

	"github.com/d5/tengo/v2"
)

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
			return WrapError(err), nil
		}

		return &tengo.String{Value: s}, nil
	}
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

// WrapError transforms the given error to a Tengo's error.
func WrapError(err error) tengo.Object {
	if err == nil {
		return tengo.TrueValue
	}
	return &tengo.Error{Value: &tengo.String{Value: err.Error()}}
}

// StringArray transforms the given params to a []string.
func StringArray(arr []tengo.Object, argName string) ([]string, error) {
	var sarr []string
	for idx, elem := range arr {
		str, ok := tengo.ToString(elem)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("%s[%d]", argName, idx),
				Expected: "string",
				Found:    elem.TypeName(),
			}
		}
		sarr = append(sarr, str)
	}
	return sarr, nil
}

// InterfaceArray transforms the given params to a []any.
func InterfaceArray(args []tengo.Object) []any {
	arguments := make([]any, len(args))
	for i, arg := range args {
		arguments[i] = tengo.ToInterface(arg)
	}
	return arguments
}

// Format executes a fmt.Sprintf on the given args, the first one must be the string `format'.
func Format(args ...tengo.Object) (string, error) {
	format, ok := tengo.ToString(args[0])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return fmt.Sprintf(format, InterfaceArray(args[1:])), nil
}
