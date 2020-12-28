package redirectschema

import (
	"net/http"
	"regexp"
	"strings"
)

type middleware struct {
	next      http.Handler
	permanent bool
	code      int
}

func New(next http.Handler, permanent bool) http.Handler {
	code := http.StatusTemporaryRedirect
	if permanent {
		code = http.StatusPermanentRedirect
	}
	return &middleware{next: next, permanent: permanent, code: code}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.TLS == nil {
		newURL := newRawURL(r)
		http.Redirect(w, r, newURL, m.code)
		return
	}
	//already is https
	m.next.ServeHTTP(w, r)
}

func newRawURL(req *http.Request) string {
	scheme := "https"
	host := req.Host
	port := ""
	uri := req.RequestURI

	schemeRegex := `^(https?):\/\/(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`
	re, _ := regexp.Compile(schemeRegex)
	if re.Match([]byte(req.RequestURI)) {
		match := re.FindStringSubmatch(req.RequestURI)
		scheme = match[1]

		if len(match[2]) > 0 {
			host = match[2]
		}

		if len(match[3]) > 0 {
			port = match[3]
		}

		uri = match[4]
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}
