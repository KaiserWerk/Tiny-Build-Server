package middleware

import (
	"golang.org/x/time/rate"
	"net/http"
)

var limiter = rate.NewLimiter(15, 30)

// Limit middleware limits the number of requests
func Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		// hmm
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
