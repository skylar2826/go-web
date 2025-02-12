package custom_error

import (
	"errors"
	"fmt"
)

var ErrorInvalidRouterPattern = func(str string) error {
	return errors.New(fmt.Sprintf("invalid router pattern: %s\n", str))
}

var ErrorShutDownServer = func(str string) error {
	return errors.New(fmt.Sprintf("shutdown server error: %s\n", str))
}

var ErrorTimeout = func(str string) error {
	return errors.New(fmt.Sprintf("timeout, err: %s\n", str))
}

var ErrorNotFound = func(str string) error {
	return errors.New(fmt.Sprintf("not found: %s\n", str))
}

var ErrorUnauthorizedJson = func(str string) error {
	return errors.New(fmt.Sprintf("unauthorized err: %s\n", str))
}
