Namecon
=======

Namecon is a simple library which provides a base for generating predefined names based on any standards provided by a type which implements the `NameGenerator` interface. This is useful for creating simple names that match needed criterias based on giving rules.

Concept
-------

Namecon is rather simple, you create a structure which implements the rules you desire for the giving standard to be used for the template provided and then use that to generate namers which will create new names based based on a provided base.

This concept is very simple but can grow into a power system which streamlines how you name instances, daemons, running apps, ...etc.

Example
-------

```go
import (
	"github.com/honeytrap/namecon"
)

func main() {

	// Create the name generator with the base template and NameGenerator rule enforcer.
	simpleNamer := namecon.GenerateNamer(namecon.SimpleNamer{}, "app-%s:%s")

	// create a new name based on a specifc group/app/subdomain.
	simpleNamer("trapper") // => app-trapper:454u34334909
	simpleNamer("trapper") // => app-trapper:u89J9232232
	simpleNamer("hippy")   // => app-hippy:fge6023ghu
}

```
