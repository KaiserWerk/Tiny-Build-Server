package middleware

import (
	"context"
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
)

func (h *MWHandler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := security.CheckLogin(h.SessSvc, r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		currentUser, err := sessionservice.GetUserFromSession(h.Ds, session)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", currentUser))

		next.ServeHTTP(w, r)
	})
}

func (h *MWHandler) AuthWithAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := security.CheckLogin(h.SessSvc, r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		currentUser, err := sessionservice.GetUserFromSession(h.Ds, session)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !currentUser.Admin {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "user", currentUser))

		next.ServeHTTP(w, r)
	})
}
