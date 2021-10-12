package organization

import (
	"github.com/giantswarm/microerror"
)

var secretNotFoundError = &microerror.Error{
	Kind: "secretNotFoundError",
}

// IsSecretNotFound asserts secretNotFoundError.
func IsSecretNotFound(err error) bool {
	return microerror.Cause(err) == secretNotFoundError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
