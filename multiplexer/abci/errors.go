package abci

import "errors"

var (
	// ErrNoVersionFound is returned when no remote version is found for a given app version.
	ErrNoVersionFound = errors.New("no version found")
)
