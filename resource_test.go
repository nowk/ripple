package ripple

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type CtrlMismatch struct {
	Namespace

	Index func(http.Handler) http.Handler `ripple:"GET /"`
}

func (CtrlMismatch) IndexFunc(_ http.ResponseWriter, _ *http.Request) {
	//
}

func TestFieldTypeDoesNotMatchMethodType(t *testing.T) {
	c := new(CtrlMismatch)

	vof := reflect.ValueOf(c).Elem()
	fld, ok := vof.Type().FieldByName("Index")
	if !ok {
		t.Fatalf("unknown field %s", "Index")
	}

	_, err := newResource(fld, vof)

	exp := fmt.Errorf("mismatched types")
	if !reflect.DeepEqual(exp, err) {
		t.Errorf("expected mismatched types error, got %s", err)
	}
}
