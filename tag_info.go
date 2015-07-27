package ripple

import (
	"errors"
	"reflect"
	"strings"
)

// tagInfo is the parts of a split ripple tag
type tagInfo []string

func (t tagInfo) Method() string {
	return t[1]
}

func (t tagInfo) Path() string {
	return strings.TrimRight(t[2], "/")
}

func (t tagInfo) Action() string {
	return t[0]
}

func (t tagInfo) Valid() bool {
	return len(t) == 3
}

var (
	errParseTagEmptyString  = errors.New("tagInfo: cannot parse empty string")
	errParseTagInvalidSplit = errors.New("tagInfo: invalid split length")
)

func newTagInfo(field reflect.StructField) (tagInfo, error) {
	tag := field.Tag.Get(fieldTagKey)
	if tag == "" {
		return nil, nil
	}

	s := strings.Split(tag, ",")
	name := s[0]
	if name == "" {
		name = field.Name
	}

	s = strings.Split(s[1], " ")

	info := tagInfo(append([]string{name}, s...))
	if !info.Valid() {
		return nil, errParseTagInvalidSplit
	}

	return info, nil
}
