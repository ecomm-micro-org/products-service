package custom_errors

import "errors"

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrNotEnoughPermissions = errors.New("you do not have enough permissions to access this resource")
)
