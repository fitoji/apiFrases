package transport

import (
	"aprende-golang/internal/model"
	"aprende-golang/internal/service"
	"aprende-golang/internal/store"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// fakeStore implementa store.Store con errores inyectables por metodo,
// para probar el mapeo de errores de DB del handler sin tocar SQLite.
type fakeStore struct {
	errGetAll    error
	errGetByID   error
	errGetRandom error
	errCrear     error
	errUpdate    error
	errDelete    error
}

var _ store.Store = (*fakeStore)(nil)

func (f *fakeStore) GetAll() ([]*model.Frase, error) { return nil, f.errGetAll }

// Devuelve un registro guardado cuando no hay error inyectado: la fachada
// store.Store no puede devolver (nil, nil) y el PUT ahora lee antes de mergear.
func (f *fakeStore) GetByID(id int) (*model.Frase, error) {
	if f.errGetByID != nil {
		return nil, f.errGetByID
	}
	return &model.Frase{
		ID:                  id,
		Frase:               "frase guardada",
		FraseIdiomaOriginal: "original guardado",
		Autor:               "autor guardado",
		Categoria:           "categoria guardada",
	}, nil
}

func (f *fakeStore) GetRandom() (*model.Frase, error) { return nil, f.errGetRandom }

func (f *fakeStore) Crear(frase *model.Frase) (*model.Frase, error) { return nil, f.errCrear }

func (f *fakeStore) Update(id int, frase *model.Frase) (*model.Frase, error) {
	return nil, f.errUpdate
}

func (f *fakeStore) Delete(id int) error { return f.errDelete }

func setupHandlerConFake(t *testing.T, f *fakeStore) *FraseHandler {
	t.Helper()
	svc := service.New(f)
	return New(svc)
}

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

// hallazgo 5: un error de DB que NO es sql.ErrNoRows debe devolver 500
// (antes todo error se disfrazaba de 404).
func TestGetFrasePorIDConErrorDeBaseDevuelve500(t *testing.T) {
	f := &fakeStore{errGetByID: errors.New("fallo de base simulado")}
	h := setupHandlerConFake(t, f)

	req := httptest.NewRequest(http.MethodGet, "/frases/999", nil)
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "error interno del servidor" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fallo de base simulado") {
		t.Fatalf("no debe filtrar el error crudo, body=%q", w.Body.String())
	}
}

// hallazgo 5: sql.ErrNoRows sigue siendo 404.
func TestGetFrasePorIDNoExistenteDevuelve404(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/frases/999", nil)
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "No lo encontramos!" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
}

// hallazgo 6: el 500 de GET /frases no debe filtrar el error crudo de DB.
func TestGetFrasesConErrorDeBaseNoFiltraError(t *testing.T) {
	f := &fakeStore{errGetAll: errors.New("fallo de base simulado")}
	h := setupHandlerConFake(t, f)

	req := httptest.NewRequest(http.MethodGet, "/frases", nil)
	w := httptest.NewRecorder()

	h.HandleFrases(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "error interno del servidor" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fallo de base simulado") {
		t.Fatalf("no debe filtrar el error crudo, body=%q", w.Body.String())
	}
}

func TestPostFrasesJSONInvalidoDevuelve400(t *testing.T) {
	h := setupHandler(t)

	body := `{"frase":`
	req := httptest.NewRequest(http.MethodPost, "/frases", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrases(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "input invalido" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
}

func TestHandleFrasesMetodoNoDisponibleDevuelve405(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/frases", nil)
	w := httptest.NewRecorder()

	h.HandleFrases(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperaba 405, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "Metodo no disponible!" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
}

// hallazgo 6: 500s de POST, random, PUT y DELETE con body seguro.
func TestPostFrasesConErrorDeBaseDevuelve500SinFiltrar(t *testing.T) {
	f := &fakeStore{errCrear: errors.New("fallo de base simulado")}
	h := setupHandlerConFake(t, f)

	body := `{"frase":"Hola","original":"Hi","autor":"Yo","categoria":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/frases", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrases(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "error interno del servidor" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fallo de base simulado") {
		t.Fatalf("no debe filtrar el error crudo, body=%q", w.Body.String())
	}
}

func TestGetFraseRandomConErrorDeBaseDevuelve500SinFiltrar(t *testing.T) {
	f := &fakeStore{errGetRandom: errors.New("fallo de base simulado")}
	h := setupHandlerConFake(t, f)

	req := httptest.NewRequest(http.MethodGet, "/frases/random", nil)
	w := httptest.NewRecorder()

	h.HandleFraseRandom(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "error interno del servidor" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fallo de base simulado") {
		t.Fatalf("no debe filtrar el error crudo, body=%q", w.Body.String())
	}
}

func TestPutFraseConErrorDeBaseDevuelve500SinFiltrar(t *testing.T) {
	f := &fakeStore{errUpdate: errors.New("fallo de base simulado")}
	h := setupHandlerConFake(t, f)

	body := `{"frase":"Hola","original":"Hi","autor":"Yo","categoria":"test"}`
	req := httptest.NewRequest(http.MethodPut, "/frases/1", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "error interno del servidor" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fallo de base simulado") {
		t.Fatalf("no debe filtrar el error crudo, body=%q", w.Body.String())
	}
}

// hallazgo 8: PUT con body parcial solo pisa los campos enviados.
// RED 1: {"frase":"nueva"} no debe blanquear original/autor/categoria.
func TestPutFraseParcialSoloFraseNoBlanqueaResto(t *testing.T) {
	h := setupHandler(t)

	creada, err := h.service.CrearFrase(model.Frase{
		Frase:               "frase original",
		FraseIdiomaOriginal: "original en ingles",
		Autor:               "autor original",
		Categoria:           "educacion",
	})
	if err != nil {
		t.Fatalf("crear frase de prueba: %v", err)
	}

	body := `{"frase":"nueva frase"}`
	req := httptest.NewRequest(http.MethodPut, "/frases/"+strconv.Itoa(creada.ID), strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d; body=%s", w.Code, w.Body.String())
	}

	guardada, err := h.service.ObtenFrasePorID(creada.ID)
	if err != nil {
		t.Fatalf("releer frase: %v", err)
	}
	if guardada.Frase != "nueva frase" {
		t.Errorf("frase: queria \"nueva frase\", obtuve %q", guardada.Frase)
	}
	if guardada.FraseIdiomaOriginal != "original en ingles" {
		t.Errorf("original: debia preservarse, obtuve %q", guardada.FraseIdiomaOriginal)
	}
	if guardada.Autor != "autor original" {
		t.Errorf("autor: debia preservarse, obtuve %q", guardada.Autor)
	}
	if guardada.Categoria != "educacion" {
		t.Errorf("categoria: debia preservarse, obtuve %q", guardada.Categoria)
	}
}

// hallazgo 8: RED 2 — {"autor":"nuevo autor"} no debe blanquear frase/categoria.
func TestPutFraseParcialSoloAutorNoBlanqueaFrase(t *testing.T) {
	h := setupHandler(t)

	creada, err := h.service.CrearFrase(model.Frase{
		Frase:               "frase original",
		FraseIdiomaOriginal: "original en ingles",
		Autor:               "autor original",
		Categoria:           "educacion",
	})
	if err != nil {
		t.Fatalf("crear frase de prueba: %v", err)
	}

	body := `{"autor":"nuevo autor"}`
	req := httptest.NewRequest(http.MethodPut, "/frases/"+strconv.Itoa(creada.ID), strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d; body=%s", w.Code, w.Body.String())
	}

	guardada, err := h.service.ObtenFrasePorID(creada.ID)
	if err != nil {
		t.Fatalf("releer frase: %v", err)
	}
	if guardada.Autor != "nuevo autor" {
		t.Errorf("autor: queria \"nuevo autor\", obtuve %q", guardada.Autor)
	}
	if guardada.Frase != "frase original" {
		t.Errorf("frase: debia preservarse, obtuve %q", guardada.Frase)
	}
	if guardada.Categoria != "educacion" {
		t.Errorf("categoria: debia preservarse, obtuve %q", guardada.Categoria)
	}
}

// hallazgo 8 (triangulación): un PUT con los cuatro campos enviados debe
// pisarlos todos — el merge no degrada la actualización completa.
func TestPutFraseCompletaActualizaTodosLosCampos(t *testing.T) {
	h := setupHandler(t)

	creada, err := h.service.CrearFrase(model.Frase{
		Frase:               "frase original",
		FraseIdiomaOriginal: "original en ingles",
		Autor:               "autor original",
		Categoria:           "educacion",
	})
	if err != nil {
		t.Fatalf("crear frase de prueba: %v", err)
	}

	body := `{"frase":"frase nueva","original":"original nuevo","autor":"autor nuevo","categoria":"filosofia"}`
	req := httptest.NewRequest(http.MethodPut, "/frases/"+strconv.Itoa(creada.ID), strings.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d; body=%s", w.Code, w.Body.String())
	}

	guardada, err := h.service.ObtenFrasePorID(creada.ID)
	if err != nil {
		t.Fatalf("releer frase: %v", err)
	}
	if guardada.Frase != "frase nueva" ||
		guardada.FraseIdiomaOriginal != "original nuevo" ||
		guardada.Autor != "autor nuevo" ||
		guardada.Categoria != "filosofia" {
		t.Errorf("los cuatro campos debian actualizarse, obtuve %+v", guardada)
	}
}

func TestDeleteFraseConErrorDeBaseDevuelve500SinFiltrar(t *testing.T) {
	f := &fakeStore{errDelete: errors.New("fallo de base simulado")}
	h := setupHandlerConFake(t, f)

	req := httptest.NewRequest(http.MethodDelete, "/frases/1", nil)
	w := httptest.NewRecorder()

	h.HandleFrasePorID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d; body=%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != "error interno del servidor" {
		t.Fatalf("body inesperado: %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fallo de base simulado") {
		t.Fatalf("no debe filtrar el error crudo, body=%q", w.Body.String())
	}
}
