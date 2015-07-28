package ripple

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func catch(fn func()) error {
	var err error
	func() {
		defer func() {
			re := recover()
			if re != nil {
				switch v := re.(type) {
				case error:
					err = v
				case string:
					err = fmt.Errorf(v)
				}
			}
		}()

		fn()
	}()

	return err
}

type setupGroupFunc func(*echo.Echo)

func send(meth, path string,
	setupGroup setupGroupFunc, t *testing.T) *httptest.ResponseRecorder {

	req, err := http.NewRequest(meth, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	echoMux := echo.New()
	setupGroup(echoMux)

	echoMux.ServeHTTP(w, req)

	return w
}
