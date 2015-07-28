package ripple

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo"
)

type CtrlBasic struct {
	Namespace

	Index  http.HandlerFunc `ripple:"GET /"`
	Create http.HandlerFunc `ripple:"POST /"`
	Show   echo.HandlerFunc `ripple:"GET /:id"`
	Update echo.HandlerFunc `ripple:"PUT /:id"`
	Del    echo.HandlerFunc `ripple:"DELETE /:id"`
}

func (CtrlBasic) IndexFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
}

func (CtrlBasic) CreateFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Create", req.Method, req.URL.Path)
}

func (CtrlBasic) ShowFunc(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, "[%s] %s #Show:%s", req.Method, req.URL.Path, c.Param("id"))
	return nil
}

func (CtrlBasic) UpdateFunc(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, "[%s] %s #Update:%s", req.Method, req.URL.Path, c.Param("id"))
	return nil
}

func (CtrlBasic) DelFunc(c *echo.Context) error {
	req := c.Request()
	c.HTML(200, "[%s] %s #Del:%s", req.Method, req.URL.Path, c.Param("id"))
	return nil
}

func TestAppliesMethodsToNewEchoGroupUsingTagsAsManifest(t *testing.T) {
	echoMux := echo.New()
	_ = Group(&CtrlBasic{Namespace: "/posts"}, echoMux)

	for _, v := range []struct {
		meth, Namespace, extra string
	}{
		{"GET", "/posts", "Index"},
		{"POST", "/posts", "Create"},
		{"GET", "/posts/123", "Show:123"},
		{"PUT", "/posts/123", "Update:123"},
		// {"PATCH", "/posts/123", "Update:123"},
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

type CtrlInternalField struct {
	Namespace

	AccessKey string
	Index     http.HandlerFunc `ripple:"GET /"`
}

func (c CtrlInternalField) IndexFunc(w http.ResponseWriter, req *http.Request) {
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

	Index http.HandlerFunc `ripple:"GET /"`
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

type CtrlAssignOnField struct {
	Namespace

	Index http.HandlerFunc `ripple:"GET /"`
}

func TestUseAssignedHandlerOnField(t *testing.T) {
	var index = func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
	}

	echoMux := echo.New()
	_ = Group(&CtrlAssignOnField{Index: index}, echoMux)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	echoMux.ServeHTTP(w, req)

	exp := "[GET] / #Index"
	got := w.Body.String()
	if exp != got {
		t.Errorf("expected %s, got %s", exp, got)
	}
}

func TestPanicWhenAssignableHandlerIsNotAssigned(t *testing.T) {
	echoMux := echo.New()

	got := catch(func() {
		_ = Group(&CtrlAssignOnField{}, echoMux)
	})

	exp := fmt.Errorf("action method not found: Index")
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected action method not found error, got %s", got.Error())
	}
}

type CtrlWithMiddleware struct {
	Namespace

	Log   echo.MiddlewareFunc `ripple:"*"`
	Index http.HandlerFunc    `ripple:"GET /"`
}

func (CtrlWithMiddleware) LogFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("log in"))
		next.ServeHTTP(w, req)
		w.Write([]byte("log out"))
	})
}

func (CtrlWithMiddleware) IndexFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
}

func TestMiddlewareSupport(t *testing.T) {
	echoMux := echo.New()
	_ = Group(&CtrlWithMiddleware{}, echoMux)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	echoMux.ServeHTTP(w, req)

	exp := "log in[GET] / #Indexlog out"
	got := w.Body.String()
	if exp != got {
		t.Errorf("expected %s, got %s", exp, got)
	}
}
