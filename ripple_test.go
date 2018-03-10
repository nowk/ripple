package ripple

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
	Group(&CtrlBasic{Namespace: ""}, echoMux)

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
		Group(&CtrlInternalField{AccessKey: "myaccesskey"}, echoMux)
	}, t)

	if got := w.Body.String(); "AccessKey is myaccesskey" != got {
		t.Errorf("incorrect handler output, got %s", got)
	}
}

func TestPanicsObjectToGroupIsNotAStruct(t *testing.T) {
	err := catch(func() {
		Group(Namespace("/posts"), echo.New())
	})

	if errNotStruct != err {
		t.Errorf("expected not struct error, got %s", err)
	}
}

type CtrlAssignOnField struct {
	Namespace

	Index http.HandlerFunc `ripple:"GET /"`
}

func TestAssignedHandlerOnConstruction(t *testing.T) {
	w := send("GET", "/", func(echoMux *echo.Echo) {
		Group(&CtrlAssignOnField{
			Index: func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(w, "[%s] %s #Index", req.Method, req.URL.Path)
			},
		}, echoMux)
	}, t)

	if got := w.Body.String(); "[GET] / #Index" != got {
		t.Errorf("incorrect handler output, got %s", got)
	}
}

func TestPanicOnUndefinedHandler(t *testing.T) {
	err := catch(func() {
		Group(&CtrlAssignOnField{}, echo.New())
	})

	if errActionNotFound("Index") != err {
		t.Errorf("expected action not found error, got %s", err)
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
		Group(&CtrlMismatch{}, echo.New())
	})

	if errTypeMismatch != err {
		t.Errorf("expected type mismatch error, got %s", err)
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
		Group(&CtrlWithMiddleware{}, echoMux)
	}, t)

	if got := w.Body.String(); "log in[GET] / #Indexlog out" != got {
		t.Errorf("expected middleware output, got %s", got)
	}
}
