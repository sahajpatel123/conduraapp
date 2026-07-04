package memory

import "errors"

var (
	// ErrNotFound is returned when a memory entry is not found.
	ErrNotFound = errors.New("memory: entry not found")

	// ErrInvalidMemoryType is returned when an invalid memory type is specified.
	ErrInvalidMemoryType = errors.New("memory: invalid memory type")

	// ErrInvalidConfidence is returned when confidence is out of range.
	ErrInvalidConfidence = errors.New("memory: confidence must be between 0.0 and 1.0")

	// ErrEmptyContent is returned when content is empty.
	ErrEmptyContent = errors.New("memory: content cannot be empty")

	// ErrStoreClosed is returned when operations are attempted on a closed store.
	ErrStoreClosed = errors.New("memory: store is closed")
)
