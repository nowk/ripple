package ripple

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type tfield struct {
	tag  string
	name string
}

func (f tfield) Tag() string {
	return f.tag
}

func (f tfield) Name() string {
	return f.name
}

func (F tfield) Type() reflect.Type {
	return reflect.TypeOf("")
}

func Test_newFieldInfoEmptyTagReturnsNilNil(t *testing.T) {
	info, err := newFieldInfo(tfield{})

	if info != nil {
		t.Error("expected info to be nil")
	}
	if err != nil {
		t.Error("expected err to be nil")
	}
}

func Test_newFieldInfoParsesTagAndReturnsAfieldInfo(t *testing.T) {
	for _, v := range []struct {
		ectype           echoType
		meth, path, name string
	}{
		{handler, "GET", "/", "Index"},
		{middleware, "*", "", "Log"},
	} {
		var tag string
		if v.meth == "*" {
			tag = v.meth

			v.meth = ""
		} else {
			tag = fmt.Sprintf("%s %s", v.meth, v.path)
		}

		tf := &tfield{
			tag:  tag,
			name: v.name,
		}
		info, err := newFieldInfo(tf)
		if err != nil {
			t.Fatal(err)
		}

		exp := &fieldInfo{
			EchoType: v.ectype,
			Path:     strings.TrimRight(v.path, "/"),
			Method:   v.meth,
			Name:     v.name,
			Type:     reflect.TypeOf(""),
		}

		if !reflect.DeepEqual(exp, info) {
			t.Errorf("expected %s to equal %s", exp, info)
		}
	}
}

func Test_newFieldErrorsOnInvalidTagFormat(t *testing.T) {
	for _, v := range []string{
		"GET/",
		"GET / ",
	} {
		_, err := newFieldInfo(&tfield{
			tag: v,
		})

		if err != errInvalidTagSplit {
			t.Error("expected invalid tag split error")
		}
	}
}
