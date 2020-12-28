package ping

import "net/http"

type pingMiddleware struct {
	next http.Handler
}

// New - create new ping instance
func New(next http.Handler) http.Handler {
	return &pingMiddleware{next: next}
}

func (p *pingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/ping" {
		p.pingHandler(w, r)
		return
	}
	p.next.ServeHTTP(w, r)
}

func (p *pingMiddleware) pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}
