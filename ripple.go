package ripple

import (
	"reflect"

	. "github.com/nowk/ripple/errors"
	"github.com/nowk/ripple/resource"
	"gopkg.in/labstack/echo.v1"
)

// Controller is the interface for a Controller to be applied to an echo Group
type Controller interface {

	// __Path is the namespace the Group at, eg /posts
	__Path() string
}

// Mount applies the Controller to the echo via a new Group using the
// Controller's ripple tags as a manifest to properly associate methods/path and
// handler.
func Mount(c Controller, ech *echo.Echo) *echo.Group {
	vof, typ, err := reflectCtrl(c)
	if err != nil {
		panic(err)
	}

	grp := ech.Group(c.__Path())

	i := 0
	j := typ.NumField()
	for ; i < j; i++ {
		res, err := resource.New(typ.Field(i), vof)
		if err != nil {
			panic(err)
		}
		if res == nil {
			continue // if there is no route
		}

		res.Apply(grp)
	}

	return grp
}

func reflectCtrl(c Controller) (reflect.Value, reflect.Type, error) {
	vof := reflect.ValueOf(c)
	typ := vof.Type()

	if typ.Kind() == reflect.Ptr {
		vof = vof.Elem()
		typ = vof.Type()
	}

	var err error
	if typ.Kind() != reflect.Struct {
		err = &Error{typ.Name(), "not a struct"}
	}

	return vof, typ, err
}

// Namespace provides an embeddable type that will allow a struct to implement
// Controller.
type Namespace string

var _ Controller = Namespace("")

// Path returns a string implementing Controller
func (n Namespace) __Path() string {
	if n == "" {
		return "/"
	}

	return string(n)
}
