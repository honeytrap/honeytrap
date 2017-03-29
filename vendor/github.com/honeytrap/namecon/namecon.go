// Package namecon defines a package for naming things, it provides us a baseline
// package for providing naming standards for things we use.
package namecon

// NameGenerator defines a type whch exposes a method which generates a new name
// based on the provided parameters.
type NameGenerator interface {
	Generate(template string, base string) string
}

// Namer defines an interface which exposes a single method to deliver a new name
// based on an internal NameGenerator and the supplied value.
type Namer interface {
	New(string) string
}

// BaseNamer defines an interface which exposes a single method to deliver a new name
// based on an internal NameGenerator.
type BaseNamer interface {
	New() string
}

//================================================================================

// NamerCon defines a struct which implements the Namer interface and uses the
// provided template and generator to generate a new name for use.
type NamerCon struct {
	template  string
	generator NameGenerator
}

// NewNamerCon returns a new instance of the NamerCon struct.
func NewNamerCon(template string, generator NameGenerator) *NamerCon {
	return &NamerCon{
		template:  template,
		generator: generator,
	}
}

// New returns a new name based on the provided base value and the internal template
// and NameGenerator.
func (n *NamerCon) New(base string) string {
	return n.generator.Generate(n.template, base)
}

//================================================================================

// Names defines a struct which implements the Namer interface and uses the
// provided template and generator to generate a new name for use.
type Names struct {
	template  string
	base      string
	generator NameGenerator
}

// NewNames returns a new instance of the NamerCon struct.
func NewNames(template string, base string, generator NameGenerator) *Names {
	return &Names{
		base:      base,
		template:  template,
		generator: generator,
	}
}

// New returns a new name based on the provided he internal template
// and generator.
func (n *Names) New() string {
	return n.generator.Generate(n.template, n.base)
}

//================================================================================

// GenerateNamer defines a function which accepts a base generator and uses the provided
// template and returns a function which will use this template and generator for all
// name generation.
func GenerateNamer(gen NameGenerator, template string) func(string) string {
	return func(base string) string {
		return gen.Generate(template, base)
	}
}
