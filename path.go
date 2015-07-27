package ripple

// Namespace provides an embeddable type that will allow a struct to implement
// Controller.
type Namespace string

var _ Controller = Namespace("")

func (n Namespace) Path() string {
	if n == "" {
		return "/"
	}

	return string(n)
}
