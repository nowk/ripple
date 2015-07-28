package ripple

import (
	"fmt"
	"github.com/labstack/echo"
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

func newResource(
	field reflect.StructField, v reflect.Value) (*resource, error) {

	info, err := newFieldInfo(structField{field})
	if err != nil || info == nil {
		return nil, err
	}

	fn, err := getResourceFunc(info, v)
	if err != nil {
		return nil, err
	}
	if !fn.Type().ConvertibleTo(info.Type) {
		return nil, fmt.Errorf("mismatched types")
	}

	return &resource{
		fieldInfo: info,

		Func: fn,
	}, nil
}

// Set sets the resources on the given group
func (r resource) Set(grp *echo.Group) {
	var (
		name string
		args []reflect.Value
	)

	if r.EchoType == middleware {
		name = "Use"
		args = append(args, r.Func)
	} else {
		name = methodMap[r.Method]
		args = append(args, reflect.ValueOf(r.Path), r.Func)
	}

	grpValue := reflect.ValueOf(grp)
	fn := grpValue.MethodByName(name)
	fn.Call(args)
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
