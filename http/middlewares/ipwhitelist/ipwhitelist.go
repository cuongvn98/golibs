// Package ipwhitelist provides ...
package ipwhitelist

import (
	"github.com/cuongvn98/golibs/ip"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type middleware struct {
	next     http.Handler
	strategy ip.Strategy
	checker  *ip.Checker
}

// New - create new ipwhitelist middleware
func New(next http.Handler, strategy ip.Strategy, sourceRange []string) (http.Handler, error) {
	if len(sourceRange) == 0 {
		return nil, errors.New("source range is empty, ipwhitelist is not created")
	}
	checker, err := ip.NewChecker(sourceRange)
	if err != nil {
		return nil, errors.Errorf("can not create ip checker with %q : %s ", strings.Join(sourceRange, ","), err.Error())
	}
	return &middleware{next: next, strategy: strategy, checker: checker}, nil
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := m.checker.IsAuthorized(m.strategy.GetIP(r))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	m.next.ServeHTTP(w, r)
}
