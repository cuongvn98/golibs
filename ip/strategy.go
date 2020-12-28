// Package ip provides ...
package ip

import (
	"net"
	"net/http"
	"strings"
)

const (
	xForwardedFor = "X-Forwarded-For"
)

// Strategy - a strategy for ip selection
type Strategy interface {
	GetIP(req *http.Request) string
}

// RemoteAddrStrategy - always return remote address
type RemoteAddrStrategy struct {
}

// GetIP - get request address
func (r *RemoteAddrStrategy) GetIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}

// DepthStrategy - a strategy based on the depth inside the X-Forwarded-For from right to left
type DepthStrategy struct {
	Depth int
}

// GetIP - return the selected IP
func (s *DepthStrategy) GetIP(req *http.Request) string {
	xff := req.Header.Get(xForwardedFor)
	xffs := strings.Split(xff, ",")
	if len(xffs) < s.Depth {
		return ""
	}
	return strings.TrimSpace(xffs[len(xffs)-s.Depth])
}
