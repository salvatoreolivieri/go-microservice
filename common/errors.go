package common

import "errors"

var (
	ErrNoItems = errors.New("items must have at least one item")
)
