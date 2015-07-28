package ripple

import (
	"errors"
	"reflect"

	"github.com/labstack/echo"
)

const fieldTagKey = "ripple"

// Controller is the interface for a Controller to be applied to an echo Group
type Controller interface {
	// Path is the namespace ripple will create the Group at, eg /posts
	Path() string
}

var errNotStruct = errors.New("invalid controller type: requires a struct type")

func reflectCtrl(c Controller) (reflect.Value, reflect.Type, error) {
	vof := reflect.ValueOf(c)
	typ := vof.Type()

	if typ.Kind() == reflect.Ptr {
		vof = vof.Elem()
		typ = vof.Type()
	}

	var err error
	if typ.Kind() != reflect.Struct {
		err = errNotStruct
	}

	return vof, typ, err
}

// Group applies the Controller to the echo via a new Group using the
// Controller's ripple tags as a manifest to properly associate methods/path and
// handler.
func Group(c Controller, echoMux *echo.Echo) *echo.Group {
	cvof, ctyp, err := reflectCtrl(c)
	if err != nil {
		panic(err)
	}

	grp := echoMux.Group(c.Path())

	i := 0
	n := ctyp.NumField()
	for ; i < n; i++ {
		re, err := newResource(ctyp.Field(i), cvof)
		if err != nil {
			panic(err)
		}
		if re == nil {
			continue // if there is no route
		}

		re.Set(grp)
	}

	return grp
}
