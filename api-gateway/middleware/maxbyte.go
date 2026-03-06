package middleware

import "net/http"

func MaxByteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 200<<20)
		next.ServeHTTP(w, r)
	})
}
