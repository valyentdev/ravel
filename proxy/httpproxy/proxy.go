package httpproxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ProxyRequest struct {
	// In is the request received by the proxy.
	// The Rewrite function must not modify In.
	In *http.Request

	// Out is the request which will be sent by the proxy.
	// The Rewrite function may modify or replace this request.
	// Hop-by-hop headers are removed from this request
	// before Rewrite is called.
	Out *http.Request
}

type FilterFunc[RCTX any] func(http.ResponseWriter, *http.Request, *RCTX) bool
type RewriteFunc[RCTX any] func(*httputil.ProxyRequest, *RCTX)

type HTTPProxyService[RCTX any] interface {
	// Request context data
	NewRCTX() *RCTX

	// Get the upstream URL to proxy the request to
	// If *url.URL is nil then a 502 Bad Gateway is sent to downstream
	GetUpstream(*http.Request, *RCTX) *url.URL
}

// Error handler to handle errors returned by ModifyResponse or errors during request processing
// defaultErrorHandler is used if nil
type WithErrorHandler[RCTX any] interface {
	ErrorHandler(http.ResponseWriter, *http.Request, *RCTX, error)
}

// Modify the response before sending it to the client
// If returns an error, the Error handler is called
// Default is to do nothing
type WithModifyResponse[RCTX any] interface {
	ModifyResponse(*http.Response) error
}

// Rewrite the request before sending it to the upstream
// useful for adding headers, etc.
// Default is to do nothing but rewrite the target and clean websocket headers
type WithRewrite[RCTX any] interface {
	Rewrite(*ProxyRequest, *RCTX)
}

// Early filter to reject / accept the request before any processing
// Should return false to reject the request
// Default is to accept all requests
type WithEarlyFilter[RCTX any] interface {
	EarlyFilter(http.ResponseWriter, *http.Request, *RCTX) bool
}

// Filter to reject / accept the request after the target is resolved
// Default is to accept all requests
type WithFilter[RCTX any] interface {
	Filter(http.ResponseWriter, *http.Request, *RCTX) bool
}

type Options struct {
	FlushInterval time.Duration
	RoundTripper  http.RoundTripper
	BufferPool    httputil.BufferPool
	ErrorLog      *log.Logger
}

func defaultFilter[RCTX any](w http.ResponseWriter, r *http.Request, rctx *RCTX) bool {
	return true
}

func defaultRewrite[RCTX any](pr *ProxyRequest, rctx *RCTX) {
}

const defaultFlushInterval = 100 * time.Millisecond

type Proxy[RCTX any] struct {
	newRCTX     func() *RCTX
	getUpstream func(*http.Request, *RCTX) *url.URL
	earlyFilter FilterFunc[RCTX]
	filter      FilterFunc[RCTX]
	rp          *httputil.ReverseProxy
}

var _ ProxyHandler = (*Proxy[any])(nil)

type proxyCtxKey int

const (
	targetKey proxyCtxKey = iota
	rctxKey
)

func withCtx[RCTX any](req *http.Request, target *url.URL, rctx *RCTX) *http.Request {
	ctx := context.WithValue(req.Context(), targetKey, target)
	ctx = context.WithValue(ctx, rctxKey, rctx)
	return req.WithContext(ctx)
}
func fromCtx[RCTX any](ctx context.Context) (*url.URL, *RCTX) {
	return ctx.Value(targetKey).(*url.URL), ctx.Value(rctxKey).(*RCTX)
}

func alwaysRewrite(pr *httputil.ProxyRequest) {
	cleanWebSocketHeaders(pr.Out)
}

// Create a new proxy from a service
// The service must implement HTTPProxyService[RCTX]
// opts may be nil
// The service can also implement the following interfaces:
// - WithErrorHandler[RCTX]
// - WithModifyResponse[RCTX]
// - WithRewrite[RCTX]
// - WithEarlyFilter[RCTX]
// - WithFilter[RCTX]
func NewProxy[RCTX any](service HTTPProxyService[RCTX], opts *Options) *Proxy[RCTX] {
	proxy := &Proxy[RCTX]{
		newRCTX:     service.NewRCTX,
		getUpstream: service.GetUpstream,
	}

	var options Options
	if opts != nil {
		options.BufferPool = opts.BufferPool
		options.ErrorLog = opts.ErrorLog
		if opts.FlushInterval == 0 {
			options.FlushInterval = defaultFlushInterval
		} else {
			options.FlushInterval = opts.FlushInterval
		}
		options.RoundTripper = opts.RoundTripper
	}

	if er, ok := service.(WithEarlyFilter[RCTX]); ok {
		proxy.earlyFilter = er.EarlyFilter
	} else {
		proxy.earlyFilter = defaultFilter[RCTX]
	}

	if fr, ok := service.(WithFilter[RCTX]); ok {
		proxy.filter = fr.Filter
	} else {
		proxy.filter = defaultFilter[RCTX]
	}

	var errorHandler func(http.ResponseWriter, *http.Request, error)
	if er, ok := service.(WithErrorHandler[RCTX]); ok {
		errorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			_, rctx := fromCtx[RCTX](r.Context())
			er.ErrorHandler(w, r, rctx, err)
		}
	} else {
		errorHandler = defaultErrorHandler
	}

	rewriteFunc := defaultRewrite[RCTX]
	if rw, ok := service.(WithRewrite[RCTX]); ok {
		rewriteFunc = rw.Rewrite
	}

	var modifyResponse func(*http.Response) error
	if mr, ok := service.(WithModifyResponse[RCTX]); ok {
		modifyResponse = mr.ModifyResponse
	}

	proxy.rp = &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			target, rctx := fromCtx[RCTX](pr.In.Context())
			rewriteTarget(target, pr)
			alwaysRewrite(pr)
			out := withCtx(pr.Out, target, rctx)
			pr.Out = out
			rewriteFunc(&ProxyRequest{
				In:  pr.In,
				Out: pr.Out,
			}, rctx)
		},
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
		Transport:      options.RoundTripper,
		FlushInterval:  options.FlushInterval,
		BufferPool:     options.BufferPool,
		ErrorLog:       options.ErrorLog,
	}
	return proxy
}

func (p *Proxy[RCTX]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rctx := p.newRCTX()

	if !p.earlyFilter(w, r, rctx) {
		return
	}

	target := p.getUpstream(r, rctx)
	if target == nil {
		AnswerErrorStatus(w, r, http.StatusBadGateway)
		return
	}

	if !p.filter(w, r, rctx) {
		return
	}

	ctx := context.WithValue(r.Context(), targetKey, target)
	ctx = context.WithValue(ctx, rctxKey, rctx)
	r = r.WithContext(ctx)

	p.rp.ServeHTTP(w, r)
}
