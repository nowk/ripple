package ripple

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo"
)

type CtrlOne struct {
	Namespace

	_ http.HandlerFunc `ripple:"GET,/,Index"`
	_ http.HandlerFunc `ripple:"POST,/,Create"`
	_ echo.HandlerFunc `ripple:"GET,/:id,Show"`
	_ echo.HandlerFunc `ripple:"PUT,/:id,Update"`
	_ echo.HandlerFunc `ripple:"PATCH,/:id,Update"`
	_ echo.HandlerFunc `ripple:"DELETE,/:id,Del"`
}

func (p CtrlOne) Index(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
}

func (p CtrlOne) Create(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Create", req.Method, req.URL.Path)
}

func (p CtrlOne) Show(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, "[%s] %s #Show:%s", req.Method, req.URL.Path, c.Param("id"))
	return nil
}

func (p CtrlOne) Update(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, "[%s] %s #Update:%s", req.Method, req.URL.Path, c.Param("id"))
	return nil
}

func (p CtrlOne) Del(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, "[%s] %s #Del:%s", req.Method, req.URL.Path, c.Param("id"))
	return nil
}

func TestAppliesMethodsToNewEchoGroupUsingTagsAsManifest(t *testing.T) {
	echoMux := echo.New()
	_ = Group(&CtrlOne{Namespace: "/posts"}, echoMux)

	for _, v := range []struct {
		meth, Namespace, extra string
	}{
		{"GET", "/posts", "Index"},
		{"POST", "/posts", "Create"},
		{"GET", "/posts/123", "Show:123"},
		{"PUT", "/posts/123", "Update:123"},
		{"PATCH", "/posts/123", "Update:123"},
		{"DELETE", "/posts/123", "Del:123"},
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

type CtrlUnknownMethod struct {
	Namespace

	_ http.HandlerFunc `ripple:"GETS,/,Index"`
}

func (CtrlUnknownMethod) Index(w http.ResponseWriter, req *http.Request) {
	//
}

func TestPanicsIfMethodIsNotInMethodMap(t *testing.T) {
	echoMux := echo.New()

	got := catch(func() {
		_ = Group(&CtrlUnknownMethod{}, echoMux)
	})

	exp := fmt.Errorf("unknown method map: %s", "GETS")
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected unknown method map error, got %s", got.Error())
	}
}

type CtrlInternalField struct {
	Namespace

	AccessKey string
	_         http.HandlerFunc `ripple:"GET,/,Index"`
}

func (c CtrlInternalField) Index(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "AccessKey is %s", c.AccessKey)
}

func TestAccessingInternalFields(t *testing.T) {
	echoMux := echo.New()
	_ = Group(&CtrlInternalField{AccessKey: "myaccesskey"}, echoMux)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	echoMux.ServeHTTP(w, req)

	exp := "AccessKey is myaccesskey"
	got := w.Body.String()
	if exp != got {
		t.Errorf("expected %s, got %s", exp, got)
	}
}

type CtrlMethodNotFound struct {
	Namespace

	_ http.HandlerFunc `ripple:"GET,/,Index"`
}

func TestPanicOnnewRouteError(t *testing.T) {
	echoMux := echo.New()

	got := catch(func() {
		_ = Group(&CtrlMethodNotFound{}, echoMux)
	})

	exp := fmt.Errorf("action method not found: Index")
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected action method not found error, got %s", got.Error())
	}
}

func TestPanicsIfNotAStruct(t *testing.T) {
	echoMux := echo.New()

	got := catch(func() {
		_ = Group(Namespace("/posts"), echoMux)
	})

	exp := errControllerInvalidType
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected %s, got %s", exp.Error(), got.Error())
	}
}
