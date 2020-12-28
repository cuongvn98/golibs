// Package recovery provides ...
package recovery

import (
	"net/http"
)

type recovery struct {
	next http.Handler
}

// New - create recovery middleware
func New(next http.Handler) http.Handler {
	return &recovery{next: next}
}

func (recovery *recovery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recovery.recoveryFunc(w, r)
	recovery.next.ServeHTTP(w, r)
}

func (recovery *recovery) recoveryFunc(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		if !shouldLogPanic(err) {
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func shouldLogPanic(panicValue interface{}) bool {
	return panicValue != nil && panicValue != http.ErrAbortHandler
}
