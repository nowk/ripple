package fieldinfo

import (
	"fmt"
	"github.com/nowk/ripple/methods"
	"reflect"
	"strings"
)

const RippleTagKey string = "ripple"

type echotype int

const (
	_ echotype = iota
	Middleware
	Handler
)

type Fieldinfo struct {
	*taginfo

	Name string
	Type reflect.Type
}

func New(f reflect.StructField) (*Fieldinfo, error) {
	ti, err := parseTag(f.Tag.Get(RippleTagKey))
	if err != nil {
		return nil, err
	}
	if ti == nil {
		return nil, nil // there was no field tag for ripple
	}

	fi := &Fieldinfo{
		taginfo: ti,

		Name: f.Name,
		Type: f.Type,
	}

	return fi, nil
}

// MethodName returns the associated method func name for ripple field.
// eg. Index -> IndexFunc
func (f *Fieldinfo) MethodName() string {
	return fmt.Sprintf("%sFunc", f.Name)
}

func (f *Fieldinfo) IsMiddleware() bool {
	return f.Echotype == Middleware
}

// taginfo is a structured representation of the the parsed field tag
type taginfo struct {
	Method   string
	Path     string
	Echotype echotype
}

func parseTag(tag string) (*taginfo, error) {
	if tag == "" {
		return nil, nil
	}
	if tag == ",middleware" {
		ti := &taginfo{
			Echotype: Middleware,
		}

		return ti, nil
	}

	split := strings.Split(tag, " ") // eg. GET /path
	if len(split) != 2 {
		return nil, fmt.Errorf("`%s`: invalid tag format", tag)
	}
	method, path := split[0], split[1]

	// check that the method is a valid HTTP method supported by echo
	_, ok := methods.Map[method]
	if !ok {
		return nil, fmt.Errorf("%s: unsupported HTTP method", method)
	}

	ti := &taginfo{
		Method: method,
		Path:   strings.TrimRight(path, "/"),

		Echotype: Handler,
	}

	return ti, nil
}
