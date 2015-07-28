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

type errActionNotFound string

func (e errActionNotFound) Error() string {
	return fmt.Sprintf("action not found: %s", string(e))
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

	return fn, errActionNotFound(info.Name)
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

func (r resource) isMiddleware() bool {
	return r.EchoType == middleware
}

func (r resource) callName() string {
	if r.isMiddleware() {
		return "Use"
	}

	return methodMap[r.Method]
}

func (r resource) callArgs() []reflect.Value {
	if r.isMiddleware() {
		return []reflect.Value{r.Func}
	}

	return []reflect.Value{
		reflect.ValueOf(r.Path),
		r.Func,
	}
}

// Set sets the resources on the given group
func (r resource) Set(grp *echo.Group) {
	fn := reflect.ValueOf(grp).MethodByName(r.callName())
	fn.Call(r.callArgs())
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
