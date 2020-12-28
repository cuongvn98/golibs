// Package forcehttps provides ...
package forcehttps

import (
	"fmt"
	"net/http"
	"strings"
)

type forceHTTPSMiddleware struct {
	next      http.Handler
	httpsPort int
}

// New - create new force https
func New(next http.Handler, httpsPort int) http.Handler {
	return &forceHTTPSMiddleware{next: next, httpsPort: httpsPort}
}

func (f *forceHTTPSMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := f.buildForceHTTPSUrl(r.Host, r.URL.RawPath, r.URL.RawQuery)
	http.Redirect(w, r, u, http.StatusMovedPermanently)
	return
}

func (f *forceHTTPSMiddleware) buildForceHTTPSUrl(host, path, query string) string {
	httpsPort := f.httpsPort
	splitStr := strings.Split(host, ":")
	var u string
	if httpsPort == 443 {
		u = fmt.Sprintf("https://%s", splitStr[0])
	} else {
		u = fmt.Sprintf("https://%s:%d", splitStr[0], httpsPort)
	}
	if path != "" && path != "/" {
		u += path
	}
	if query != "" {
		u += "?" + query
	}
	return u
}
