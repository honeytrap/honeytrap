package providers

type Provider interface {
	NewContainer(string) (Container, error)
}
