package httpproxy

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

	"golang.org/x/net/http/httpguts"
)

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

func GetStatusCode(err error) int {
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

func StatusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}

func AnswerErrorStatus(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(StatusText(statusCode))); err != nil {
		slog.Error("Failed to write response", "error", err)
	}
}

type ProxyHandler http.Handler

func rewriteTarget(target *url.URL, pr *httputil.ProxyRequest) {
	out := pr.Out
	out.URL.Scheme = target.Scheme
	out.URL.Host = target.Host

	u := out.URL
	if out.RequestURI != "" {
		parsedURL, err := url.ParseRequestURI(out.RequestURI)
		if err == nil {
			u = parsedURL
		}
	}

	out.URL.Path = u.Path
	out.URL.RawPath = u.RawPath
	out.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
	out.RequestURI = "" // Outgoing request should not have RequestURI
}

func defaultErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
	statusCode := GetStatusCode(err)
	AnswerErrorStatus(w, req, statusCode)
}

func StripHostPort(h string) string {
	// If no port on host, return unchanged
	if !strings.Contains(h, ":") {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h // on error, return unchanged
	}
	return host
}
