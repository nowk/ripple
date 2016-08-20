package errors

import (
	"fmt"
)

type Error struct {
	Desc string
	Msg  string
}

func (e *Error) Error() string {
	if e.Desc != "" {
		return fmt.Sprintf("%s: %s", e.Desc, e.Msg)
	}

	return e.Msg
}
