package ripple

import (
	"fmt"
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
