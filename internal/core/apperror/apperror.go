package apperror

import "fmt"

type (
	// UserError is a general error type for the user-facing issues.
	// You can still use this as a container for more specific errors.
	UserError struct {
		ErrorGetting
		ErrorUpdating
		ErrorDeleting
		DuplicateError
		NotFound
		InvalidLoginCredentials
		InvalidResource
		CustomError
		Message string
	}

	// ErrorGetting represents an error when fetching a resource.
	ErrorGetting struct {
		Resource string
	}

	// CustomError represents a generic custom error.
	CustomError struct {
		Message string
	}

	// InvalidResource represents an invalid resource error.
	InvalidResource struct {
		Resource string
	}

	// ErrorUpdating represents an error when updating a resource.
	ErrorUpdating struct {
		Resource string
	}

	// ErrorDeleting represents an error when deleting a resource.
	ErrorDeleting struct {
		Resource string
	}

	// DuplicateError represents a duplicate resource error.
	DuplicateError struct {
		Resource string
	}

	// InvalidLoginCredentials represents an error for incorrect login details.
	InvalidLoginCredentials struct{}

	// ErrorProcessing represents a general error during processing.
	ErrorProcessing struct {
		Action   string
		Resource string
	}

	// NotFound represents a "resource not found" error.
	NotFound struct {
		Resource string
	}
)

func (e UserError) Error() string {
	return e.Message
}

func (e ErrorGetting) Error() string {
	return fmt.Sprintf("unable to get %s at this time", e.Resource)
}

func (e CustomError) Error() string {
	return e.Message
}

func (e ErrorUpdating) Error() string {
	return fmt.Sprintf("unable to update %s at this time", e.Resource)
}

func (e ErrorDeleting) Error() string {
	return fmt.Sprintf("unable to delete %s at this time", e.Resource)
}

func (e DuplicateError) Error() string {
	return fmt.Sprintf("%s already exists", e.Resource)
}

func (e ErrorProcessing) Error() string {
	return fmt.Sprintf("unable to process %s at this time", e.Action)
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e InvalidResource) Error() string {
	return fmt.Sprintf("your %s is invalid", e.Resource)
}

func (e InvalidLoginCredentials) Error() string {
	return fmt.Sprintf("invalid email or password")
}
