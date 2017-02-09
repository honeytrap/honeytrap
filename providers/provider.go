package providers

import "github.com/op/go-logging"

var log = logging.MustGetLogger("honeytrap:providers")

type Provider interface {
	NewContainer(string) (Container, error)
}
