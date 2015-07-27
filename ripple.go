package ripple

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/labstack/echo"
)

const fieldTagKey = "ripple"

// Controller is the interface for a Controller to be applied to an echo Group
type Controller interface {
	// Path is the namespace ripple will create the Group at, eg /posts
	Path() string
}

// methodMap maps all echo methods that match the func(string, echo.Handler)
// signature used to add method routes
var methodMap = map[string]string{
	"GET":     "Get",
	"POST":    "Post",
	"PUT":     "Put",
	"PATCH":   "Patch",
	"DELETE":  "Delete",
	"HEAD":    "Head",
	"OPTIONS": "Options",
	"CONNECT": "Connect",
	"TRACE":   "Trace",

	// TODO add WebSocket?
}

var errControllerInvalidType = errors.New("controllers must be struct types")

func reflectCtrl(c Controller) (reflect.Value, reflect.Type) {
	vof := reflect.ValueOf(c)
	typ := vof.Type()

	if typ.Kind() == reflect.Ptr {
		vof = vof.Elem()
		typ = vof.Type()
	}

	if typ.Kind() != reflect.Struct {
		panic(errControllerInvalidType)
	}

	return vof, typ
}

// Group applies the Controller to the echo via a new Group using the
// Controller's ripple tags as a manifest to properly associate methods/path and
// handler.
func Group(c Controller, echoMux *echo.Echo) *echo.Group {
	var (
		grp      = echoMux.Group(c.Path())
		grpValue = reflect.ValueOf(grp)

		cValue, cType = reflectCtrl(c)
	)

	i := 0
	n := cType.NumField()
	for ; i < n; i++ {
		rt, err := newRoute(cValue, cType.Field(i))
		if err != nil {
			panic(err)
		}
		if rt == nil {
			continue // if there is no route
		}

		meth, ok := methodMap[rt.Method]
		if !ok {
			panic(fmt.Errorf("unknown method map: %s", rt.Method))
		}
		methFunc := grpValue.MethodByName(meth)
		methFunc.Call(rt.CallArgs())
	}

	return grp
}
