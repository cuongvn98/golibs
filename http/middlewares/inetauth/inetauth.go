// Package inetauth provides ...
package inetauth

import (
	"context"
	"net/http"
	"strings"
)

type contextUserKey string

// USER - key of user payload in request context
const USER contextUserKey = "USER"

type middleware struct {
	next         http.Handler
	validUsers   []string
	removeHeader bool
}

// New - create new inetauth middleware
func New(next http.Handler, rawUsers string, ignoreCredential bool) http.Handler {
	users := getLinesFromRaw(rawUsers)
	return &middleware{next: next, validUsers: users, removeHeader: ignoreCredential}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")

	errf := func() {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	if err != nil {
		errf()
		return
	}
	u, err := exchange(cookie.Value)
	if err != nil {
		errf()
		return
	}
	if m.removeHeader {
		r.AddCookie(&http.Cookie{
			Name:  "sid",
			Value: "",
		})
	}
	if isValidUser(m.validUsers, u.Email) {
		ctx := context.WithValue(r.Context(), USER, u)
		r = r.WithContext(ctx)
		m.next.ServeHTTP(w, r)
		return
	}
	errf()
}

func isValidUser(userEmails []string, email string) bool {
	if len(userEmails) == 0 {
		return true
	}
	for _, v := range userEmails {
		if v == email {
			return true
		}
	}
	return false
}

func getLinesFromRaw(rawUsers string) []string {
	var filteredLines []string
	for _, rawLine := range strings.Split(rawUsers, "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" && !strings.HasPrefix(line, "#") {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}
