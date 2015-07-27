package ripple

import (
	"fmt"
	"reflect"
)

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
	tInfo, err := newTagInfo(field)
	if err != nil {
		return nil, err
	}
	if tInfo == nil {
		return nil, nil // no ripple tag
	}

	handler, err := getHandler(tInfo.ActionName(), v)
	if err != nil {
		return nil, err
	}

	// TODO check that the field type matches the method signature

	return &route{
		Method:  tInfo.Method(),
		Path:    tInfo.Path(),
		Handler: handler,
	}, nil
}

func (r route) ToArgs() []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(r.Path),
		r.Handler,
	}
}
