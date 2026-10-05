package middleware

import "net/http"

// CORS returns an HTTP middleware that adds permissive CORS headers to every
// response for routes under /api/. This is intentionally open for local
// development; production deployment (out of scope for v1) would restrict
// the allowed origin to the specific frontend host.
//
// The middleware handles preflight OPTIONS requests directly and allows
// the methods and headers used by the StudentHub frontend (api.js).
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		// Respond to preflight requests immediately without passing to the
		// next handler — the real handler would reject the OPTIONS method.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
