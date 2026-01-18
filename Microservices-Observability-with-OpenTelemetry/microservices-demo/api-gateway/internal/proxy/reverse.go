package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New(prefix, target string) http.Handler {
	u, _ := url.Parse(target)
	rp := httputil.NewSingleHostReverseProxy(u)

	return http.StripPrefix(
		prefix,
		otelhttp.NewHandler(rp, prefix),
	)
}
