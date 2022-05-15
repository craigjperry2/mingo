package errors

import "errors"

var ErrUnknownUser = errors.New("unable to determine username")
var ErrUnknownHost = errors.New("unable to determine hostname")
var ErrPortUnavailable = errors.New("unable to bind on port")
