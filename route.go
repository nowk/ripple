package ripple

import (
	"fmt"
	"reflect"
)

type resource struct {
	*fieldInfo

	Func reflect.Value // TODO do we have any need to make this echo.Handler?
}

func getResourceFunc(info *fieldInfo, v reflect.Value) (reflect.Value, error) {
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

func newResource(v reflect.Value, field reflect.StructField) (*resource, error) {
	info, err := newFieldInfo(structField{field})
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil // no ripple tag
	}

	fn, err := getResourceFunc(info, v)
	if err != nil {
		return nil, err
	}

	// TODO check that the field type matches the method signature

	return &resource{
		fieldInfo: info,

		Func: fn,
	}, nil
}

func (r resource) CallName() string {
	if r.EchoType == middleware {
		return "Use"
	}

	return methodMap[r.Method]
}

func (r resource) CallArgs() []reflect.Value {
	args := []reflect.Value{r.Func}

	if r.EchoType == middleware {
		return args
	}

	return append([]reflect.Value{reflect.ValueOf(r.Path)}, args...)
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
