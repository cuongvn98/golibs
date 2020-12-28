package forwardedheader

import (
	"github.com/hirosume/libs/ip"
	"net"
	"net/http"
	"strings"
)

const (
	xForwardedProto             = "X-Forwarded-Proto"
	xForwardedFor               = "X-Forwarded-For"
	xForwardedHost              = "X-Forwarded-Host"
	xForwardedPort              = "X-Forwarded-Port"
	xForwardedServer            = "X-Forwarded-Server"
	xForwardedURI               = "X-Forwarded-Uri"
	xForwardedMethod            = "X-Forwarded-Method"
	xForwardedTLSClientCert     = "X-Forwarded-Tls-Client-Cert"
	xForwardedTLSClientCertInfo = "X-Forwarded-Tls-Client-Cert-Info"
	xRealIP                     = "X-Real-Ip"
	connection                  = "Connection"
	upgrade                     = "Upgrade"
)

var xHeaders = []string{
	xForwardedProto,
	xForwardedFor,
	xForwardedHost,
	xForwardedPort,
	xForwardedServer,
	xForwardedURI,
	xForwardedMethod,
	xForwardedTLSClientCert,
	xForwardedTLSClientCertInfo,
	xRealIP,
}

type middleware struct {
	next       http.Handler
	insecure   bool
	trustedIps []string
	ipChecker  *ip.Checker
}

func New(next http.Handler, insecure bool, trustedIps []string) (http.Handler, error) {
	var ipChecker *ip.Checker
	if len(trustedIps) > 0 {
		var err error
		ipChecker, err = ip.NewChecker(trustedIps)
		if err != nil {
			return nil, err
		}
	}

	return &middleware{next: next, insecure: insecure, trustedIps: trustedIps, ipChecker: ipChecker}, nil
}

func (m *middleware) isTrustedIP(ip string) bool {
	if m.ipChecker == nil {
		return false
	}
	return m.ipChecker.IsAuthorized(ip) != nil
}

// removeIPv6Zone removes the zone if the given IP is an ipv6 address and it has {zone} information in it,
// like "[fe80::d806:a55d:eb1b:49cc%vEthernet (vmxnet3 Ethernet Adapter - Virtual Switch)]:64692".
func removeIPv6Zone(clientIP string) string {
	return strings.Split(clientIP, "%")[0]
}

func isWebsocketRequest(req *http.Request) bool {
	containsHeader := func(name, value string) bool {
		items := strings.Split(req.Header.Get(name), ",")
		for _, item := range items {
			if value == strings.ToLower(strings.TrimSpace(item)) {
				return true
			}
		}
		return false
	}
	return containsHeader(connection, "upgrade") && containsHeader(upgrade, "websocket")
}

func forwardedPort(req *http.Request) string {
	if req == nil {
		return ""
	}

	if _, port, err := net.SplitHostPort(req.Host); err == nil && port != "" {
		return port
	}
	if req.Header.Get(xForwardedProto) == "https" || req.Header.Get(xForwardedProto) == "wss" {
		return "443"
	}
	if req.TLS != nil {
		return "443"
	}
	return "80"
}
func (m *middleware) rewrite(outreq *http.Request) {
	if clientIP, _, err := net.SplitHostPort(outreq.RemoteAddr); err != nil {
		clientIP := removeIPv6Zone(clientIP)

		if outreq.Header.Get(xRealIP) == "" {
			outreq.Header.Set(xRealIP, clientIP)
		}
	}

	xProto := outreq.Header.Get(xForwardedProto)

	if xProto == "" {
		if isWebsocketRequest(outreq) {
			if outreq.TLS != nil {
				outreq.Header.Set(xForwardedProto, "wss")
			} else {
				outreq.Header.Set(xForwardedProto, "ws")
			}
		} else {
			if outreq.TLS != nil {
				outreq.Header.Set(xForwardedProto, "https")
			} else {
				outreq.Header.Set(xForwardedProto, "http")
			}
		}
	}

	if xfPort := outreq.Header.Get(xForwardedPort); xfPort == "" {
		outreq.Header.Set(xForwardedPort, forwardedPort(outreq))
	}

	if xfHost := outreq.Header.Get(xForwardedHost); xfHost == "" && outreq.Host != "" {
		outreq.Header.Set(xForwardedHost, outreq.Host)
	}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !m.insecure && !m.isTrustedIP(r.RemoteAddr) {
		for _, header := range xHeaders {
			r.Header.Del(header)
		}
	}

	m.rewrite(r)

	m.next.ServeHTTP(w, r)
}
