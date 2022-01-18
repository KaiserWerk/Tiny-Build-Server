package middleware

import (
	"fmt"
	"net/http"
)

func (h *MWHandler) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("recover handler: %v", r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
