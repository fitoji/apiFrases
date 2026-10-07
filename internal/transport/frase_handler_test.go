package transport

import (
	"aprende-golang/internal/service"
	"aprende-golang/internal/store"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestDBTransport(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("abrir base en memoria: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	const ddl = `
		CREATE TABLE IF NOT EXISTS frases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			frase TEXT NOT NULL,
			original TEXT NOT NULL,
			autor TEXT NOT NULL,
			categoria TEXT NOT NULL
		)
	`
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("crear tabla: %v", err)
	}
	return db
}

func setupHandler(t *testing.T) *FraseHandler {
	t.Helper()
	db := newTestDBTransport(t)
	s := store.New(db)
	svc := service.New(s)
	return New(svc)
}

func TestPostFrasesSinFraseDevuelve400(t *testing.T) {
	h := setupHandler(t)

	body := `{"frase":""}`
	req := httptest.NewRequest(http.MethodPost, "/frases", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrases(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "necesitamos una frase" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
}

func TestPutFraseSinFraseDevuelve400(t *testing.T) {
	h := setupHandler(t)

	body := `{"frase":""}`
	req := httptest.NewRequest(http.MethodPut, "/frases/1", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "necesitamos una frase" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
}

func TestPutFraseIdNoExistenteDevuelve404(t *testing.T) {
	h := setupHandler(t)

	body := `{"frase":"Hola","original":"Hi","autor":"Yo","categoria":"test"}`
	req := httptest.NewRequest(http.MethodPut, "/frases/999", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "No lo encontramos!" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
}
