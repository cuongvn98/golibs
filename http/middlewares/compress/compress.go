// Package compress provides ...
package compress

import (
	"compress/gzip"
	"fmt"
	"mime"
	"net/http"

	"github.com/NYTimes/gziphandler"
)

type compress struct {
	next     http.Handler
	name     string
	excludes []string
}

// New - create new compress instance
func New(next http.Handler, excluses []string) (http.Handler, error) {
	excludes := []string{"application/grpc"}

	for _, v := range excluses {
		mediaType, _, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, fmt.Errorf("compress: cant parse exlude mimetype: %s %s", v, err.Error())
		}
		excludes = append(excludes, mediaType)
	}
	return &compress{next: next, excludes: excludes}, nil
}

func (c *compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	//if err != nil {
	//log.("compress: %s", err.Error())
	//}
	if contains(c.excludes, mediaType) {
		c.next.ServeHTTP(w, r)
	} else {
		gzipHandler(c.next).ServeHTTP(w, r)
	}
}

func gzipHandler(h http.Handler) http.Handler {
	wrapper, err := gziphandler.GzipHandlerWithOpts(
		gziphandler.CompressionLevel(gzip.DefaultCompression),
		gziphandler.MinSize(gziphandler.DefaultMinSize),
	)
	if err != nil {
		//logger.Debugf("compress gziphandler: %s", err.Error())
	}
	return wrapper(h)
}

func contains(values []string, val string) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}
	return false
}
