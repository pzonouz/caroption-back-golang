package middlewares

import (
	"net/http"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func AdminOrReadOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)

			return
		}

		user := utils.GetUserFromRequest(w, r)
		if user.IsAdmin {
			next.ServeHTTP(w, r)

			return
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := utils.GetUserFromRequest(w, r)
		if user.IsAdmin {
			next.ServeHTTP(w, r)

			return
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}
	})
}
