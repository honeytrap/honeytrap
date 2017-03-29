package namecon

import "fmt"

var (
	maxCharacter      = 60
	maxBaseCharacter  = 10
	maxAvailableSpace = maxCharacter - maxBaseCharacter
)

// SimpleNamer defines a struct which implements the NameGenerator interface and generates
// a giving name based on the template and a giving set of directives.
type SimpleNamer struct{}

// Generate returns a new name from the provided template and base to generate based
// on the rules of the simple rules.
func (s SimpleNamer) Generate(template string, base string) string {
	if len(base) > maxBaseCharacter {
		base = base[:maxBaseCharacter]
	}

	gened := fmt.Sprintf(template, base, String(maxBaseCharacter))

	if len(gened) > maxCharacter {
		gened = gened[:maxCharacter]
	}

	return gened
}

//================================================================================

// Basic defines a struct which implements the NameGenerator interface and generates
// a giving name based on the template and a giving set of directives.
type Basic struct{}

// Generate returns a new name based on the provided arguments.
func (Basic) Generate(template string, base string) string {
	return fmt.Sprintf(template, base)
}

//================================================================================

// LimitedNamer defines a struct which implements the NameGenerator interface and generates
// a giving name based on the template and a giving set of directives.
type LimitedNamer struct {
	maxLen     int
	maxBaseLen int
}

// NewLimitNamer returns a new instance of a LimitedNamer.
func NewLimitNamer(maxChars int, maxbaseChar int) *LimitedNamer {
	return &LimitedNamer{
		maxLen:     maxChars,
		maxBaseLen: maxbaseChar,
	}
}

// Generate returns a new name from the provided template and base to generate based
// on the rules of the simple rules.
func (l LimitedNamer) Generate(template string, base string) string {
	if len(base) > l.maxBaseLen {
		base = base[:l.maxBaseLen]
	}

	rem := l.maxLen - l.maxBaseLen
	gened := fmt.Sprintf(template, base, String(rem))

	if len(gened) > l.maxLen {
		gened = gened[:l.maxLen]
	}

	return gened
}
