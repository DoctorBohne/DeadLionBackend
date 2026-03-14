package custom_errors

import "errors"

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

var ErrAlreadBoarded = errors.New("already boarded")
