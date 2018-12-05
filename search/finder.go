package search

import "github.com/jcmturner/dependency/components"

type Finder interface {
	Find(srcRoot string) ([]components.Component, error) // Find should walk the source root to find dependency management configurations and process them to understand the dependencies.
	Class() components.Class
	Type() components.Type // Type returns the type of dependency this finder identifies
}
