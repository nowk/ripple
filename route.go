package ripple

import (
	"fmt"
	"reflect"
)

type route struct {
	*fieldInfo

	Handler reflect.Value // TODO do we have any need to make this echo.Handler?
}

func getHandler(info *fieldInfo, v reflect.Value) (reflect.Value, error) {
	var fn reflect.Value

	// first search methods
	fn = v.MethodByName(info.MethodName())
	if fn.IsValid() {
		return fn, nil
	}

	// then search fields
	fn = v.FieldByName(info.Name)
	if fn.IsValid() && !reflect.ValueOf(fn.Interface()).IsNil() {
		return fn, nil
	}

	return fn, fmt.Errorf("action method not found: %s", info.Name)
}

func newRoute(v reflect.Value, field reflect.StructField) (*route, error) {
	info, err := newFieldInfo(structField{field})
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil // no ripple tag
	}

	handler, err := getHandler(info, v)
	if err != nil {
		return nil, err
	}

	// TODO check that the field type matches the method signature

	return &route{
		fieldInfo: info,

		Handler: handler,
	}, nil
}

func (r route) CallArgs() []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(r.Path),
		r.Handler,
	}
}

// structField is a wrapper that implements structFielder
type structField struct {
	field reflect.StructField
}

func (f structField) Tag() string {
	return f.field.Tag.Get(fieldTagKey)
}

func (f structField) Name() string {
	return f.field.Name
}

func (f structField) Type() reflect.Type {
	return f.field.Type
}
