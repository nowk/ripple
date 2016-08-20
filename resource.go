package ripple

import (
	"reflect"

	. "github.com/nowk/ripple/errors"
	"github.com/nowk/ripple/fieldinfo"
	"github.com/nowk/ripple/methods"
	"gopkg.in/labstack/echo.v1"
)

// resource represents the handler/middleware to be mounted onto an Echo Group
type resource struct {
	*fieldinfo.Fieldinfo

	ControllerName string
	Func           reflect.Value
}

func newResource(f reflect.StructField, ctrl reflect.Value) (*resource, error) {
	fi, err := fieldinfo.New(f)
	if err != nil {
		return nil, err
	}
	if fi == nil {
		return nil, nil // no ripple tag
	}

	fn, err := getFunc(fi, ctrl)
	if err != nil {
		return nil, err
	}
	res := &resource{
		Fieldinfo: fi,

		ControllerName: ctrl.Type().Name(),
		Func:           fn,
	}

	return res, nil
}

// getFunc returns the associated <name>Func method for a defined ripple
// field or the actual field value if the <name>Func association is not found.
func getFunc(fi *fieldinfo.Fieldinfo, v reflect.Value) (reflect.Value, error) {
	// first search methods
	if fn := v.MethodByName(fi.MethodName()); fn.IsValid() {
		return checkFunc(fn, fi)
	}

	// then search fields
	if fn := v.FieldByName(fi.Name); fn.IsValid() && !fn.IsNil() {
		return checkFunc(fn, fi)
	}

	return reflect.Value{}, &Error{fi.Name, "action not found"}
}

// checkFunc checks to ensure the func found is convertable to the type defined
// in the field
func checkFunc(
	fn reflect.Value, fi *fieldinfo.Fieldinfo) (reflect.Value, error) {

	if !fn.Type().ConvertibleTo(fi.Type) {
		return fn, &Error{fi.Name, "type mismatch"}
	}

	return fn, nil
}

func (r resource) callName() string {
	if r.IsMiddleware() {
		return "Use"
	}

	return methods.Map[r.Method]
}

// Set sets the resources on the given group
func (r resource) Set(grp *echo.Group) {
	reflect.
		ValueOf(grp).
		MethodByName(r.callName()).
		Call(r.callArgs())
}

func (r resource) callArgs() []reflect.Value {
	if r.IsMiddleware() {
		return []reflect.Value{r.Func}
	}

	handlerFunc, ok := r.Func.Interface().(func(*echo.Context) error)
	if !ok {
		return []reflect.Value{
			reflect.ValueOf(r.Path),
			r.Func,
		}
	}

	var (
		controller_name = r.ControllerName
		action_name     = r.Name
	)
	fn := echo.HandlerFunc(func(ctx *echo.Context) error {
		ctx.Set("__controller_name", controller_name)
		ctx.Set("__action_name", action_name)

		return handlerFunc(ctx)
	})

	return []reflect.Value{
		reflect.ValueOf(r.Path),
		reflect.ValueOf(fn),
	}
}
