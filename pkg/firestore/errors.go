package firestore

import (
	"fmt"
)

// DocumentNotFoundError is the error returned when a document is not found
type DocumentNotFoundError struct {
	path string
}

func (e *DocumentNotFoundError) Error() string {
	return fmt.Sprintf("The document at path: %s does not exist", e.path)
}
