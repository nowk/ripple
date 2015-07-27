package ripple

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type tagInfo []string

func (t tagInfo) Method() string {
	return t[1]
}

func (t tagInfo) Path() string {
	return t[2]
}

func (t tagInfo) Action() string {
	return t[0]
}

func (t tagInfo) Valid() bool {
	return len(t) == 3
}

var (
	errParseTagEmptyString = errors.New(
		"parseRippleTag: cannot parse empty string")

	errParseTagInvalidSplit = errors.New("parseRippleTag: invalid split length")
)

func parseActionName(tagStr string, fieldName string) (string, string) {
	s := strings.Split(tagStr, ",")
	name := s[0]
	if name == "" {
		name = fieldName
	}

	return name, s[1]
}

func parseMethPath(str string) (string, string) {
	s := strings.Split(str, " ")
	return s[0], s[1]
}

func parseRippleTag(field reflect.StructField) (tagInfo, error) {
	tag := field.Tag.Get(fieldTagKey)
	if tag == "" {
		return nil, nil
	}

	var info tagInfo

	name, methPath := parseActionName(tag, field.Name)
	info = append(info, name)
	meth, path := parseMethPath(methPath)
	info = append(info, meth, path)

	if !info.Valid() {
		return nil, errParseTagInvalidSplit
	}

	return info, nil
}

func trimSlashR(path string) string {
	return strings.TrimRight(path, "/")
}

type route struct {
	Method  string
	Path    string
	Handler reflect.Value // TODO do we have any need to make this echo.Handler?
}

func getHandler(name string, v reflect.Value) (reflect.Value, error) {
	var fn reflect.Value

	// first search methods
	fn = v.MethodByName(name)
	if fn.IsValid() {
		return fn, nil
	}

	// then search fields
	fn = v.FieldByName(name)
	if fn.IsValid() && !reflect.ValueOf(fn.Interface()).IsNil() {
		return fn, nil
	}

	return fn, fmt.Errorf("action method not found: %s", name)
}

func newRoute(v reflect.Value, field reflect.StructField) (*route, error) {
	tInfo, err := parseRippleTag(field)
	if err != nil {
		return nil, err
	}
	if tInfo == nil {
		return nil, nil // no ripple tag
	}

	handler, err := getHandler(tInfo.Action(), v)
	if err != nil {
		return nil, err
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
