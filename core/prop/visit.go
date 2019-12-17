package prop

// Interface to implement for callers to react to a property structure traversal.
type Visitor interface {
	// Returns true if property should be visited; if false, the property will not be visited.
	ShouldVisit(property Property) bool
	// Visit the property, only when ShouldVisit returns true. If this method returns non-nil error,
	// the rest of the traversal will be aborted.
	Visit(property Property) error
	// Invoked when the children properties of a container property is about to be visited. The containing
	// property is supplied as an argument to provide context information. The container property itself,
	// however, has already been invoked on ShouldVisit and/or Visit.
	BeginChildren(container Container)
	// Invoked when the children properties of a container property has finished. The containing property
	// is supplied as a context argument.
	EndChildren(container Container)
}

// Entry point to visit a property in a depth-first-search fashion.
func Visit(property Property, visitor Visitor) error {
	if !visitor.ShouldVisit(property) {
		return nil
	}

	if err := visitor.Visit(property); err != nil {
		return err
	}

	if container, ok := property.(Container); ok {
		visitor.BeginChildren(container)
		if err := container.ForEachChild(func(_ int, child Property) error {
			return Visit(child, visitor)
		}); err != nil {
			return err
		}
		visitor.EndChildren(container)
	}

	return nil
}
