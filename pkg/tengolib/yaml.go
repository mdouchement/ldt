package tengolib

import (
	"github.com/d5/tengo/v2"
	"github.com/mdouchement/ldt/pkg/tengolib/yaml"
)

var yamlModule = map[string]tengo.Object{
	"decode": &tengo.UserFunction{
		Name:  "decode",
		Value: yamlDecode,
	},
	"encode": &tengo.UserFunction{
		Name:  "encode",
		Value: yamlEncode,
	},
}

func yamlDecode(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	switch o := args[0].(type) {
	case *tengo.Bytes:
		v, err := yaml.Decode(o.Value)
		if err != nil {
			return &tengo.Error{
				Value: &tengo.String{Value: err.Error()},
			}, nil
		}
		return v, nil
	case *tengo.String:
		v, err := yaml.Decode([]byte(o.Value))
		if err != nil {
			return &tengo.Error{
				Value: &tengo.String{Value: err.Error()},
			}, nil
		}
		return v, nil
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "bytes/string",
			Found:    args[0].TypeName(),
		}
	}
}

func yamlEncode(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	b, err := yaml.Encode(args[0])
	if err != nil {
		return &tengo.Error{Value: &tengo.String{Value: err.Error()}}, nil
	}

	return &tengo.Bytes{Value: b}, nil
}
