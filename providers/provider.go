package providers

// Provider defines a function which returns a NewContainer.
type Provider interface {
	NewContainer(string) (Container, error)
}
