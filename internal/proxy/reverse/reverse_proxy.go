package proxy

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"
)

const (
	RavelInstanceIDHeader = "X-Ravel-Instance-Id"
)

func newProxy(target *url.URL, instanceId string) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite:       func(pr *httputil.ProxyRequest) { rewrite(target, pr, instanceId) },
		FlushInterval: 100 * time.Millisecond,
		ErrorHandler:  errorHandler,
	}
}

func rewrite(target *url.URL, pr *httputil.ProxyRequest, instanceId string) {
	out := pr.Out
	out.URL.Scheme = target.Scheme
	out.URL.Host = target.Host

	out.Header.Set(RavelInstanceIDHeader, instanceId)

	u := out.URL
	if out.RequestURI != "" {
		parsedURL, err := url.ParseRequestURI(out.RequestURI)
		if err == nil {
			u = parsedURL
		}
	}

	out.URL.Path = u.Path
	out.URL.RawPath = u.RawPath
	// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
	out.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
	out.RequestURI = "" // Outgoing request should not have RequestURI

	out.Proto = "HTTP/1.1"
	out.ProtoMajor = 1
	out.ProtoMinor = 1

	cleanWebSocketHeaders(out)

}

// cleanWebSocketHeaders Even if the websocket RFC says that headers should be case-insensitive,
// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
// https://tools.ietf.org/html/rfc6455#page-20
func cleanWebSocketHeaders(req *http.Request) {
	if !isWebSocketUpgrade(req) {
		return
	}

	req.Header["Sec-WebSocket-Key"] = req.Header["Sec-Websocket-Key"]
	delete(req.Header, "Sec-Websocket-Key")

	req.Header["Sec-WebSocket-Extensions"] = req.Header["Sec-Websocket-Extensions"]
	delete(req.Header, "Sec-Websocket-Extensions")

	req.Header["Sec-WebSocket-Accept"] = req.Header["Sec-Websocket-Accept"]
	delete(req.Header, "Sec-Websocket-Accept")

	req.Header["Sec-WebSocket-Protocol"] = req.Header["Sec-Websocket-Protocol"]
	delete(req.Header, "Sec-Websocket-Protocol")

	req.Header["Sec-WebSocket-Version"] = req.Header["Sec-Websocket-Version"]
	delete(req.Header, "Sec-Websocket-Version")
}

func isWebSocketUpgrade(req *http.Request) bool {
	return httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		strings.EqualFold(req.Header.Get("Upgrade"), "websocket")
}

func errorHandler(w http.ResponseWriter, req *http.Request, err error) {
	statusCode := getStatusCode(err)
	w.WriteHeader(statusCode)
	if _, werr := w.Write([]byte(statusText(statusCode))); werr != nil {
		slog.Error("Failed to write response", "error", werr)
	}
}

func getStatusCode(err error) int {
	switch {
	case errors.Is(err, io.EOF):
		return http.StatusBadGateway
	case errors.Is(err, context.Canceled):
		return StatusClientClosedRequest
	default:
		var netErr net.Error
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				return http.StatusGatewayTimeout
			}

			return http.StatusBadGateway
		}
	}

	return http.StatusInternalServerError
}

// StatusClientClosedRequest is a custom status code to indicate that the client closed the request.
const StatusClientClosedRequest = 499
const StatusClientClosedRequestText = "Client Closed Request"

func statusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}
