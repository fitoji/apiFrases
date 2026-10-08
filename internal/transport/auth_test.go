package transport

import (
	"net/http"
	"testing"
)

const clavePrueba = "secreta-123"

// Las lecturas (GET) y el preflight (OPTIONS) son públicos: sin header deben
// llegar al handler subyacente con 200.
func TestGetSinAPIKeyPasaPublico(t *testing.T) {
	called := false
	h := WithAuth(clavePrueba, stub(&called))
	rec := request(h, http.MethodGet, "/frases", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if !called {
		t.Error("el GET público debe alcanzar el handler")
	}
}

func TestOptionsSinAPIKeyPasaPublico(t *testing.T) {
	called := false
	h := WithAuth(clavePrueba, stub(&called))
	rec := request(h, http.MethodOptions, "/frases", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if !called {
		t.Error("el OPTIONS debe alcanzar el handler")
	}
}

// Escrituras sin header X-API-Key → 401 y el handler no se ejecuta.
func TestPostSinHeaderResponde401(t *testing.T) {
	called := false
	h := WithAuth(clavePrueba, stub(&called))
	rec := request(h, http.MethodPost, "/frases", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d", rec.Code)
	}
	if called {
		t.Error("el handler no debe alcanzarse sin API key")
	}
}

// Escrituras con clave incorrecta → 401, sin filtrar detalles al cliente.
func TestPostConClaveIncorrectaResponde401(t *testing.T) {
	called := false
	h := WithAuth(clavePrueba, stub(&called))
	rec := request(h, http.MethodPost, "/frases", map[string]string{
		"X-API-Key": "otra-clave",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d", rec.Code)
	}
	if called {
		t.Error("el handler no debe alcanzarse con clave incorrecta")
	}
}

// Escrituras con la clave correcta → 200 (el stub responde) y handler alcanzado.
func TestPostConClaveCorrectaResponde200(t *testing.T) {
	called := false
	h := WithAuth(clavePrueba, stub(&called))
	rec := request(h, http.MethodPost, "/frases", map[string]string{
		"X-API-Key": clavePrueba,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if !called {
		t.Error("con la clave correcta el handler debe procesar la escritura")
	}
}

// Fail-closed: API sin API_KEY configurada no acepta escrituras, ni con header.
func TestPostConConfigVaciaResponde401(t *testing.T) {
	called := false
	h := WithAuth("", stub(&called))
	rec := request(h, http.MethodPost, "/frases", map[string]string{
		"X-API-Key": "",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d", rec.Code)
	}
	if called {
		t.Error("API sin clave configurada no debe aceptar escrituras")
	}
}

// El cuerpo del 401 es fijo: no filtra headers, claves ni errores internos.
func TestCuerpo401EsFijoSinFugas(t *testing.T) {
	called := false
	h := WithAuth(clavePrueba, stub(&called))
	rec := request(h, http.MethodPut, "/frases/1", map[string]string{
		"X-API-Key": "clave-mala-que-no-debe-aparecer",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d", rec.Code)
	}
	if got, want := rec.Body.String(), "credenciales invalidas\n"; got != want {
		t.Errorf("cuerpo: esperaba %q, obtuve %q", want, got)
	}
}
