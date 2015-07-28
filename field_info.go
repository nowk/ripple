package ripple

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// echoType represents one of the 2 types that can be mounted onto an Echo Group
// either an Handler or a Middlerware
type echoType int

const (
	_ echoType = iota

	middleware
	handler
)

// fieldInfo is the basic meta data parsed from a struct field. This does not
// include the actual field value or the <name>Func method it represents.
type fieldInfo struct {
	Path   string
	Method string
	Name   string
	Type   reflect.Type

	// EchoType represents the type of field in relation to echo either a handler
	// or middleware
	EchoType echoType
}

type structFielder interface {
	Tag() string
	Name() string
	Type() reflect.Type
}

var errTagFormat = errors.New("invalid tag format")

// methodMap maps all echo methods that match the func(string, echo.Handler)
// signature used to add method routes
var methodMap = map[string]string{
	"GET":     "Get",
	"POST":    "Post",
	"PUT":     "Put",
	"PATCH":   "Patch",
	"DELETE":  "Delete",
	"HEAD":    "Head",
	"OPTIONS": "Options",
	"CONNECT": "Connect",
	"TRACE":   "Trace",

	// TODO add WebSocket?
}

type errHttpMethod string

func (e errHttpMethod) Error() string {
	return fmt.Sprintf("invalid HTTP method: %s", string(e))
}

func parseTag(str string) (string, string, error) {
	split := strings.Split(str, " ")
	if len(split) != 2 {
		return "", "", errTagFormat
	}

	meth := split[0]
	path := split[1]

	_, ok := methodMap[meth]
	if !ok {
		return "", "", errHttpMethod(meth)
	}

	return meth, path, nil
}

func newFieldInfo(f structFielder) (*fieldInfo, error) {
	tag := f.Tag()
	if tag == "" {
		return nil, nil
	}

	var (
		err error

		ecType     echoType
		meth, path string
	)

	if tag != "*" {
		meth, path, err = parseTag(tag)
		if err != nil {
			return nil, err
		}

		ecType = handler
	} else {
		ecType = middleware
	}

	return &fieldInfo{
		Method: meth,
		Path:   strings.TrimRight(path, "/"),
		Name:   f.Name(),
		Type:   f.Type(),

		EchoType: ecType,
	}, nil
}

func (f fieldInfo) MethodName() string {
	return fmt.Sprintf("%sFunc", f.Name)
}
