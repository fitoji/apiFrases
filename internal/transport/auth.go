package transport

import (
	"crypto/subtle"
	"log"
	"net/http"
)

// WithAuth envuelve un handler con autenticación por API key (hallazgo 12).
// GET y OPTIONS pasan sin verificar (lecturas públicas; el preflight lo
// intercepta WithCors antes de llegar aquí). El resto de métodos (POST/PUT/
// DELETE) exige la cabecera X-API-Key igual a apiKey. apiKey vacía = fail-closed:
// una API sin clave configurada no acepta escrituras. En denegación se loguea
// el detalle en servidor y se responde 401 con cuerpo fijo, sin filtrar
// claves ni errores internos al cliente.
func WithAuth(apiKey string, next http.Handler) http.Handler {
	key := []byte(apiKey)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		if !keyMatches(key, r.Header.Get("X-API-Key")) {
			log.Printf("acceso denegado: %s %s desde %s", r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, "credenciales invalidas", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// keyMatches compara en tiempo constante la clave configurada con la
// recibida. Las longitudes se chequean aparte (y las vacías se descartan
// antes) para no depender del short-circuit de ConstantTimeCompare.
func keyMatches(configured []byte, provided string) bool {
	got := []byte(provided)
	if len(configured) == 0 || len(got) == 0 || len(configured) != len(got) {
		return false
	}
	return subtle.ConstantTimeCompare(configured, got) == 1
}
