package ripple

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type tagInfo []string

func (t tagInfo) Method() string {
	return t[0]
}

func (t tagInfo) Path() string {
	return t[1]
}

func (t tagInfo) Action() string {
	return t[2]
}

func (t tagInfo) Valid() bool {
	return len(t) == 3
}

var (
	errParseTagEmptyString  = errors.New("parseTag: cannot parse empty string")
	errParseTagInvalidSplit = errors.New("parstTag: invalid split length")
)

func parseTag(tag string) (tagInfo, error) {
	if tag == "" {
		return nil, errParseTagEmptyString
	}

	tInfo := tagInfo(strings.Split(tag, ","))
	if !tInfo.Valid() {
		return nil, errParseTagInvalidSplit
	}

	return tInfo, nil
}

func trimSlashR(path string) string {
	return strings.TrimRight(path, "/")
}

type route struct {
	Method  string
	Path    string
	Handler reflect.Value // TODO do we have any need to make this echo.Handler?
}

func newRoute(v reflect.Value, field reflect.StructField) (*route, error) {
	tag := field.Tag.Get(fieldTagKey)
	if tag == "" {
		return nil, nil
	}

	tInfo, err := parseTag(tag)
	if err != nil {
		return nil, err
	}

	handler := v.MethodByName(tInfo.Action())
	if !handler.IsValid() {
		return nil, fmt.Errorf("action method not found: %s", tInfo.Action())
	}

	// TODO check that the field type matches the method signature

	return &route{
		Method:  tInfo.Method(),
		Path:    trimSlashR(tInfo.Path()),
		Handler: handler,
	}, nil
}

func (r route) ToArgs() []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(r.Path),
		r.Handler,
	}
}
