// Package buffer provides ...
package buffer

import (
	"fmt"
	"net/http"

	oxybuffer "github.com/vulcand/oxy/buffer"
)

// var (
// 	mB int64 = 10 * 1024 * 1024
// )

type middleware struct {
	buffer *oxybuffer.Buffer
}

var (
	//defaullt 2MB
	memReq = oxybuffer.MemRequestBodyBytes(2 << 20)
	memRes = oxybuffer.MemResponseBodyBytes(2 << 20)
)

// New - create buffer middleware
func New(next http.Handler, maxReq, maxRes int64) (http.Handler, error) {
	oxyBuffer, err := oxybuffer.New(
		next,
		memReq,
		memRes,
	)

	switch {
	case maxReq != 0 && maxRes == 0:
		oxyBuffer, err = oxybuffer.New(
			next,
			memReq,
			oxybuffer.MaxRequestBodyBytes(maxReq),
			memRes,
		)
	case maxReq == 0 && maxRes != 0:
		oxyBuffer, err = oxybuffer.New(
			next,
			memReq,
			memRes,
			oxybuffer.MaxResponseBodyBytes(maxRes),
		)
	case maxReq != 0 && maxRes != 0:
		oxyBuffer, err = oxybuffer.New(
			next,
			memReq,
			oxybuffer.MaxRequestBodyBytes(maxReq),
			memRes,
			oxybuffer.MaxResponseBodyBytes(maxRes),
		)

	}

	if err != nil {
		return nil, fmt.Errorf("cant not create buffer middleware: %v", err)
	}
	return &middleware{buffer: oxyBuffer}, nil
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.buffer.ServeHTTP(w, r)
}
