package resource

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

func New(f reflect.StructField, ctrl reflect.Value) (*resource, error) {
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

	// then search fields to see if field was defined, eg Index: func(...)
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

// Apply sets the resources on the given group
func (r *resource) Apply(grp *echo.Group) {
	vofGrp := reflect.ValueOf(grp)
	fn := vofGrp.MethodByName(r.methodName())
	fn.Call(r.args())
}

// methodName returns the echo method name to call to bind the current resource
func (r *resource) methodName() string {
	if r.IsMiddleware() {
		return "Use"
	}

	return methods.Map[r.Method]
}

func (r *resource) args() []reflect.Value {
	if r.IsMiddleware() {
		return []reflect.Value{r.Func}
	}

	vofPath := reflect.ValueOf(r.Path)
	vofFunc := r.Func
	// type assertion to get the actual func, so we can call it directly later
	// on
	handler, ok := r.Func.Interface().(func(*echo.Context) error)
	if ok {
		vofFunc = reflect.ValueOf(actionFunc(handler, r.ControllerName, r.Name))
	}

	return []reflect.Value{
		vofPath,
		vofFunc,
	}
}

func actionFunc(
	handler echo.HandlerFunc, controllerName, actionName string) echo.HandlerFunc {

	return echo.HandlerFunc(func(ctx *echo.Context) error {
		ctx.Set("__controller_name", controllerName)
		ctx.Set("__action_name", actionName)

		return handler(ctx)
	})

}
