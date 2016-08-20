package ripple

import (
	"errors"
	"fmt"
	"reflect"

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

func newResource(f reflect.StructField, v reflect.Value) (*resource, error) {

	fieldinf, err := fieldinfo.New(f)
	if err != nil {
		return nil, err
	}
	if fieldinf == nil {
		return nil, nil
	}

	fn, err := getResourceFunc(fieldinf, v)
	if err != nil {
		return nil, err
	}
	if !fn.Type().ConvertibleTo(fieldinf.Type) {
		return nil, errTypeMismatch
	}

	return &resource{
		Fieldinfo: fieldinf,

		ControllerName: v.Type().Name(),
		Func:           fn,
	}, nil
}

func (r resource) isMiddleware() bool {
	return r.Echotype == fieldinfo.Middleware
}

func (r resource) callName() string {
	if r.isMiddleware() {
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
	if r.isMiddleware() {
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

var errTypeMismatch = errors.New("field and method types do not match")

// getResourceFunc returns the associated <name>Func method for a defined ripple
// field or the actual field value if the <name>Func association is not found.
func getResourceFunc(
	fieldinf *fieldinfo.Fieldinfo, v reflect.Value) (reflect.Value, error) {

	var fn reflect.Value

	// first search methods
	fn = v.MethodByName(fieldinf.MethodName())
	if fn.IsValid() {
		return fn, nil
	}

	// then search fields
	fn = v.FieldByName(fieldinf.Name)
	if fn.IsValid() && !reflect.ValueOf(fn.Interface()).IsNil() {
		return fn, nil
	}

	return fn, errActionNotFound(fieldinf.Name)
}

type errActionNotFound string

func (e errActionNotFound) Error() string {
	return fmt.Sprintf("action not found: %s", string(e))
}
