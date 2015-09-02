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

// structFielder is the basic interface we need for a struct field
type structFielder interface {
	Tag() string
	Name() string
	Type() reflect.Type
}

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

func newFieldInfo(f structFielder) (*fieldInfo, error) {
	taginf, err := parseTag(f.Tag())
	if err != nil {
		return nil, err
	}
	if taginf == nil {
		return nil, nil
	}

	return &fieldInfo{
		Method: taginf.meth,
		Path:   strings.TrimRight(taginf.path, "/"),
		Name:   f.Name(),
		Type:   f.Type(),

		EchoType: taginf.EchoType,
	}, nil
}

// MethodName returns the associated method name for ripple field.
// eg. Index -> IndexFunc
func (f fieldInfo) MethodName() string {
	return fmt.Sprintf("%sFunc", f.Name)
}

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

// tagInfo represents the decoded tag string
type tagInfo struct {
	meth, path string

	EchoType echoType
}

func parseTag(tag string) (*tagInfo, error) {
	if tag == "" {
		return nil, nil
	}
	if tag == ",middleware" {
		return &tagInfo{EchoType: middleware}, nil
	}

	arr := strings.Split(tag, " ")
	if len(arr) != 2 {
		return nil, errTagFormat
	}

	meth := arr[0]
	path := arr[1]

	_, ok := methodMap[meth]
	if !ok {
		return nil, errHttpMethod(meth)
	}

	return &tagInfo{meth, path, handler}, nil
}

var errTagFormat = errors.New("invalid tag format")

type errHttpMethod string

func (e errHttpMethod) Error() string {
	return fmt.Sprintf("invalid HTTP method: %s", string(e))
}
