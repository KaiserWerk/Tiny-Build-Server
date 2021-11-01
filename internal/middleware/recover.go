package middleware

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"net/http"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				logging.GetLoggerWithContext("recoverMiddleware").Errorf("%v", r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
