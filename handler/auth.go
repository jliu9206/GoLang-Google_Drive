package handler

import "net/http"

func HTTPInterceptor(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			/////////////// Interceptor Logic /////////////////////
			// parse form
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")
			// validate token
			if isTokenValid := IsTokenValid(username, token); !isTokenValid || len(username) < 3 {
				http.Error(w, "Forbidden", http.StatusUnauthorized)
				return
			}
			//////////////////////////////////////////////////////////

			// if ok
			next(w, r)
		},
	)
}
