package transport

import (
	"net/http"
	"strings"
)

// WithCors envuelve un handler con soporte CORS.
// allowedOrigins es una lista separada por comas de orígenes permitidos
// (se recortan espacios). Vacío o en blanco = cualquier origen con "*".
// El preflight (OPTIONS + Access-Control-Request-Method) se responde aquí
// con 204 y nunca llega al router.
func WithCors(allowedOrigins string, next http.Handler) http.Handler {
	allowed := parseOrigins(allowedOrigins) // nil = modo comodín
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowed != nil {
			// Las respuestas varían según Origin: marcarlo siempre en modo allowlist.
			w.Header().Add("Vary", "Origin")
		}

		origin := r.Header.Get("Origin")
		if origin != "" && (allowed == nil || allowed[origin]) {
			if allowed == nil {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		// Preflight: cortar antes del router.
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// parseOrigins convierte la lista CSV en set; devuelve nil si no hay
// orígenes válidos (modo comodín).
func parseOrigins(csv string) map[string]bool {
	set := make(map[string]bool)
	for _, o := range strings.Split(csv, ",") {
		if o = strings.TrimSpace(o); o != "" {
			set[o] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}
