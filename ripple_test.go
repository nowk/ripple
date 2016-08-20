package ripple

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/nowk/ripple/errors"
	"gopkg.in/labstack/echo.v1"
)

type CtrlBasic struct {
	Namespace

	Index  http.HandlerFunc `ripple:"GET /"`
	Create http.HandlerFunc `ripple:"POST /"`
	Show   echo.HandlerFunc `ripple:"GET :id"`
	Update echo.HandlerFunc `ripple:"PUT :id"`
	Del    echo.HandlerFunc `ripple:"DELETE :id"`
}

func (CtrlBasic) IndexFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
}

func (CtrlBasic) CreateFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Create", req.Method, req.URL.Path)
}

func (CtrlBasic) ShowFunc(c *echo.Context) error {
	req := c.Request()
	c.Request()
	c.HTML(200, fmt.Sprintf("[%s] %s #Show:%s", req.Method, req.URL.Path, c.Param("id")))
	return nil
}

func (CtrlBasic) UpdateFunc(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, fmt.Sprintf("[%s] %s #Update:%s", req.Method, req.URL.Path, c.Param("id")))
	return nil
}

func (CtrlBasic) DelFunc(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, fmt.Sprintf("[%s] %s #Del:%s", req.Method, req.URL.Path, c.Param("id")))
	return nil
}

func TestAppliesMethodsToNewEchoGroupUsingTagsAsManifest(t *testing.T) {
	echoMux := echo.New()
	Mount(&CtrlBasic{Namespace: ""}, echoMux)

	for _, v := range []struct {
		meth, Namespace, extra string
	}{
		{"GET", "/", "Index"},
		{"POST", "/", "Create"},
		{"GET", "/123", "Show:123"},
		{"PUT", "/123", "Update:123"},
		// {"PATCH", "/posts/123", "Update:123"},
		{"DELETE", "/123", "Del:123"},
	} {
		req, err := http.NewRequest(v.meth, v.Namespace, nil)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()

		echoMux.ServeHTTP(w, req)

		exp := fmt.Sprintf("[%s] %s #%s", v.meth, v.Namespace, v.extra)
		got := w.Body.String()
		if exp != got {
			t.Errorf("expected %s, got %s", exp, got)
		}
	}
}

type CtrlInternalField struct {
	Namespace

	AccessKey string
	Index     http.HandlerFunc `ripple:"GET /"`
}

func (c CtrlInternalField) IndexFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "AccessKey is %s", c.AccessKey)
}

func TestAccessingInternalFields(t *testing.T) {
	w := send("GET", "/", func(echoMux *echo.Echo) {
		Mount(&CtrlInternalField{AccessKey: "myaccesskey"}, echoMux)
	}, t)

	if got := w.Body.String(); "AccessKey is myaccesskey" != got {
		t.Errorf("incorrect handler output, got %s", got)
	}
}

type CtrlMethodNotFound struct {
	Namespace

	Index http.HandlerFunc `ripple:"GET /"`
}

func TestPanicOnnewRouteError(t *testing.T) {
	err := catch(func() {
		Mount(&CtrlMethodNotFound{}, echo.New())
	})

	var (
		exp = &Error{"Index", "action not found"}
		got = err
	)
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected %v, got %v", exp, got)
	}
}

func TestPanicsIfNotAStruct(t *testing.T) {
	err := catch(func() {
		Mount(Namespace("/posts"), echo.New())
	})

	var (
		exp = &Error{"Namespace", "not a struct"}
		got = err
	)
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected %v, got %v", exp, got)
	}
}

type CtrlAssignOnField struct {
	Namespace

	Index http.HandlerFunc `ripple:"GET /"`
}

func TestUseAssignedHandlerOnField(t *testing.T) {
	w := send("GET", "/", func(echoMux *echo.Echo) {
		Mount(&CtrlAssignOnField{
			Index: func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
			},
		}, echoMux)
	}, t)

	if got := w.Body.String(); "[GET] / #Index" != got {
		t.Errorf("incorrect handler output, got %s", got)
	}
}

func TestPanicWhenAssignableHandlerIsNotAssigned(t *testing.T) {
	err := catch(func() {
		Mount(&CtrlAssignOnField{}, echo.New())
	})

	var (
		exp = &Error{"Index", "action not found"}
		got = err
	)
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected %v, got %v", exp, got)
	}
}

type CtrlMismatch struct {
	Namespace

	Index func(http.Handler) http.Handler `ripple:"GET /"`
}

func (CtrlMismatch) IndexFunc(_ http.ResponseWriter, _ *http.Request) {
	//
}

func TestFieldTypeDoesNotMatchMethodType(t *testing.T) {
	err := catch(func() {
		Mount(&CtrlMismatch{}, echo.New())
	})

	var (
		exp = &Error{"Index", "type mismatch"}
		got = err
	)
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected %v, got %v", exp, got)
	}
}

type CtrlWithMiddleware struct {
	Namespace

	Log   echo.MiddlewareFunc `ripple:",middleware"`
	Index http.HandlerFunc    `ripple:"GET /"`
}

func (CtrlWithMiddleware) LogFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(ctx *echo.Context) error {
		ctx.Response().Write([]byte("log in"))
		err := next(ctx)
		if err != nil {
			return err
		}
		ctx.Response().Write([]byte("log out"))

		return nil
	})
}

func (CtrlWithMiddleware) IndexFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
}

func TestMiddlewareSupport(t *testing.T) {
	w := send("GET", "/", func(echoMux *echo.Echo) {
		Mount(&CtrlWithMiddleware{}, echoMux)
	}, t)

	if got := w.Body.String(); "log in[GET] / #Indexlog out" != got {
		t.Errorf("expected middleware output, got %s", got)
	}
}

type CtrlMeta struct {
	Namespace

	Index echo.HandlerFunc `ripple:"GET /"`
}

func (c CtrlMeta) IndexFunc(ctx *echo.Context) error {
	var (
		controller_name = ctx.Get("__controller_name")
		action_name     = ctx.Get("__action_name")
	)

	return ctx.HTML(200, fmt.Sprintf("%s:%s", controller_name, action_name))
}

func TestRouteMetaDataIsContextedOnRequest(t *testing.T) {
	ech := echo.New()
	Mount(&CtrlMeta{}, ech)

	var cases = []struct {
		method string
		path   string

		controller_name string
		action_name     string
	}{
		{"GET", "/", "CtrlMeta", "Index"},
	}

	for _, v := range cases {
		req, err := http.NewRequest(v.method, v.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()

		ech.ServeHTTP(w, req)

		var (
			exp = fmt.Sprintf("%s:%s", v.controller_name, v.action_name)
			got = w.Body.String()
		)
		if exp != got {
			t.Errorf("expected %s, got %s", exp, got)
		}
	}
}
