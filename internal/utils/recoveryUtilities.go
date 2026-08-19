package utils

import (
	"runtime/debug"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
)

// SafeGo starts a new goroutine with panic recovery
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log the error and stack trace
				logging.Error("Recovered from panic in goroutine: ", r)
				debug.PrintStack()
			}
		}()
		fn()
	}()
}
