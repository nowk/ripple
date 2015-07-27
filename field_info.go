package ripple

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type echoType int

const (
	_ echoType = iota

	middleware
	handler
)

// fieldInfo is the basic meta data parsed from a struct field
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

var errInvalidTagSplit = errors.New("invalid tag split")

func parseTag(str string) (string, string, error) {
	s := strings.Split(str, " ")
	if len(s) != 2 {
		return "", "", errInvalidTagSplit
	}

	return s[0], s[1], nil
}

func newFieldInfo(f structFielder) (*fieldInfo, error) {
	tag := f.Tag()
	if tag == "" {
		return nil, nil
	}

	var (
		err error

		ectype     echoType
		meth, path string
	)

	if tag != "*" {
		meth, path, err = parseTag(tag)
		if err != nil {
			return nil, err
		}

		ectype = handler
	} else {
		ectype = middleware
	}

	return &fieldInfo{
		Method: meth,
		Path:   strings.TrimRight(path, "/"),
		Name:   f.Name(),
		Type:   f.Type(),

		EchoType: ectype,
	}, nil
}

func (f fieldInfo) MethodName() string {
	return fmt.Sprintf("%sFunc", f.Name)
}
