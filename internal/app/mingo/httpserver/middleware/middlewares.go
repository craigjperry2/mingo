package middleware


import "net/http"

type middleware func(http.Handler) http.Handler

type Middlewares []middleware

// Wrap handler with each listed middleware, in order
func (mws Middlewares) Apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].Apply(mws[0](hdlr))
}
