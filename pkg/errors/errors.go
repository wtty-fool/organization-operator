package errors

import "fmt"

// NamespaceCreationError is returned when a namespace cannot be created
type NamespaceCreationError struct {
	Name string
	Err  error
}

func (e *NamespaceCreationError) Error() string {
	return fmt.Sprintf("failed to create namespace %s: %v", e.Name, e.Err)
}

// NamespaceDeletionError is returned when a namespace cannot be deleted
type NamespaceDeletionError struct {
	Name string
	Err  error
}

func (e *NamespaceDeletionError) Error() string {
	return fmt.Sprintf("failed to delete namespace %s: %v", e.Name, e.Err)
}
