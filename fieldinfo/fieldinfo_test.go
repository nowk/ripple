package fieldinfo

import (
	"reflect"
	"testing"

	"gopkg.in/labstack/echo.v1"
)

type ctrl struct {
	Middleware echo.Middleware  `ripple:",middleware"`
	Action     echo.HandlerFunc `ripple:"GET /:id"`
}

func TestReturnsFieldInfoOfGivenFieldsRippleTag(t *testing.T) {
	valueof := reflect.ValueOf(ctrl{})
	typ := valueof.Type()

	{
		fi, err := New(typ.Field(0))
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		var (
			exp = &Fieldinfo{
				taginfo: &taginfo{
					Path:     "",
					Method:   "",
					Echotype: Middleware,
				},

				Name: "Middleware",
				Type: typ.Field(0).Type,
			}

			got = fi
		)
		if !reflect.DeepEqual(exp, got) {
			t.Errorf("expected %v, got %v", exp, got)
		}
	}

	{
		fi, err := New(typ.Field(1))
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		var (
			exp = &Fieldinfo{
				taginfo: &taginfo{
					Path:     "/:id",
					Method:   "GET",
					Echotype: Handler,
				},

				Name: "Action",
				Type: typ.Field(1).Type,
			}

			got = fi
		)
		if !reflect.DeepEqual(exp, got) {
			t.Errorf("expected %v, got %v", exp, got)
		}
	}
}
