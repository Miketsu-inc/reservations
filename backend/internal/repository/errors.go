package repository

import "errors"

var ErrNotFound = errors.New("not found")
var ErrNotUnique = errors.New("not uniqe")
var ErrConstraintViolation = errors.New("constraint violation")
