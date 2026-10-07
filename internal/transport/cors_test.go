package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stub devuelve un handler que marca `called` y responde 200 "ok".
func stub(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func request(h http.Handler, method, target string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

var originDev = "http://localhost:5173"

func TestPreflightSinAllowlistResponde204(t *testing.T) {
	called := false
	h := WithCors("", stub(&called))
	rec := request(h, http.MethodOptions, "/frases", map[string]string{
		"Origin":                         originDev,
		"Access-Control-Request-Method":  http.MethodPost,
		"Access-Control-Request-Headers": "Content-Type",
	})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("esperaba 204, obtuve %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("ACAO: esperaba *, obtuve %q", got)
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), http.MethodPost) {
		t.Errorf("ACAM sin POST: %q", rec.Header().Get("Access-Control-Allow-Methods"))
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Errorf("ACAH sin Content-Type: %q", rec.Header().Get("Access-Control-Allow-Headers"))
	}
	if rec.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("falta Access-Control-Max-Age")
	}
	if called {
		t.Error("el preflight no debe llegar al handler")
	}
}

func TestGetConOriginPorDefectoAgregaCabeceras(t *testing.T) {
	called := false
	h := WithCors("", stub(&called))
	rec := request(h, http.MethodGet, "/frases", map[string]string{"Origin": originDev})

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("ACAO: esperaba *, obtuve %q", got)
	}
	if !called {
		t.Error("el handler debe procesar el GET")
	}
}

func TestSinOriginNoAgregaCabeceras(t *testing.T) {
	called := false
	h := WithCors("", stub(&called))
	rec := request(h, http.MethodGet, "/frases", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("no debe haber ACAO sin Origin, obtuve %q", got)
	}
	if !called {
		t.Error("el handler debe procesar el GET")
	}
}

func TestAllowlistOrigenPermitidoHaceEcho(t *testing.T) {
	called := false
	h := WithCors("https://app.ejemplo.com", stub(&called))
	rec := request(h, http.MethodGet, "/frases", map[string]string{"Origin": "https://app.ejemplo.com"})

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.ejemplo.com" {
		t.Errorf("ACAO: esperaba el origen exacto, obtuve %q", got)
	}
	if !strings.Contains(rec.Header().Get("Vary"), "Origin") {
		t.Errorf("Vary debe contener Origin, obtuve %q", rec.Header().Get("Vary"))
	}
	if !called {
		t.Error("el handler debe procesar el GET")
	}
}

func TestAllowlistOrigenDenegadoNoAgregaACAO(t *testing.T) {
	called := false
	h := WithCors("https://app.ejemplo.com", stub(&called))
	rec := request(h, http.MethodGet, "/frases", map[string]string{"Origin": "https://otro.com"})

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("origen denegado no debe recibir ACAO, obtuve %q", got)
	}
	if !strings.Contains(rec.Header().Get("Vary"), "Origin") {
		t.Errorf("Vary debe contener Origin también en modo allowlist, obtuve %q", rec.Header().Get("Vary"))
	}
	if rec.Code != http.StatusOK {
		t.Errorf("el handler igual debe correr (CORS no es control de acceso): esperaba 200, obtuve %d", rec.Code)
	}
	if !called {
		t.Error("el handler debe procesar el GET")
	}
}

func TestPreflightOrigenDenegadoResponde204SinACAO(t *testing.T) {
	called := false
	h := WithCors("https://app.ejemplo.com", stub(&called))
	rec := request(h, http.MethodOptions, "/frases", map[string]string{
		"Origin":                         "https://otro.com",
		"Access-Control-Request-Method":  http.MethodPost,
		"Access-Control-Request-Headers": "Content-Type",
	})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("esperaba 204, obtuve %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("preflight denegado no debe recibir ACAO, obtuve %q", got)
	}
	if called {
		t.Error("el preflight no debe llegar al handler")
	}
}

func TestAllowlistToleraEspaciosYComas(t *testing.T) {
	called := false
	h := WithCors(" https://a.com , https://b.com ", stub(&called))
	rec := request(h, http.MethodGet, "/frases", map[string]string{"Origin": "https://b.com"})

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://b.com" {
		t.Errorf("ACAO: esperaba https://b.com, obtuve %q", got)
	}
}

func TestOptionsSinPreflightPasaAlHandler(t *testing.T) {
	called := false
	h := WithCors("", stub(&called))
	rec := request(h, http.MethodOptions, "/frases", map[string]string{"Origin": originDev})

	if !called {
		t.Error("OPTIONS sin Access-Control-Request-Method debe llegar al router (que responde 405)")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("esperaba 200 del stub, obtuve %d", rec.Code)
	}
}
