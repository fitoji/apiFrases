package store

import (
	"aprende-golang/internal/model"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// newTestDB abre SQLite en memoria y crea la tabla con el DDL de SQLite.
func newTestDB(t *testing.T) *sql.DB {
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

// nuevaFrase arma una frase de prueba sin ID.
func nuevaFrase(frase, original, autor, categoria string) *model.Frase {
	return &model.Frase{
		Frase:               frase,
		FraseIdiomaOriginal: original,
		Autor:               autor,
		Categoria:           categoria,
	}
}

func TestCrearDevuelveIDYPersiste(t *testing.T) {
	s := New(newTestDB(t))

	creada, err := s.Crear(nuevaFrase("Somos lo que pensamos.", "Manopubbaṅgamā", "Buda", "budismo"))
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}
	if creada.ID <= 0 {
		t.Errorf("esperaba ID > 0, obtuve %d", creada.ID)
	}

	obtenida, err := s.GetByID(creada.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if obtenida.Frase != "Somos lo que pensamos." {
		t.Errorf("frase: esperaba %q, obtuve %q", "Somos lo que pensamos.", obtenida.Frase)
	}
	if obtenida.Autor != "Buda" || obtenida.Categoria != "budismo" {
		t.Errorf("autor/categoria: obtuve %q/%q", obtenida.Autor, obtenida.Categoria)
	}
}

func TestGetAllDevuelveLoCreado(t *testing.T) {
	s := New(newTestDB(t))

	if _, err := s.Crear(nuevaFrase("Uno", "o1", "A", "c1")); err != nil {
		t.Fatalf("Crear 1: %v", err)
	}
	if _, err := s.Crear(nuevaFrase("Dos", "o2", "B", "c2")); err != nil {
		t.Fatalf("Crear 2: %v", err)
	}

	todas, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(todas) != 2 {
		t.Fatalf("esperaba 2 frases, obtuve %d", len(todas))
	}
}

func TestGetByIDNoExistenteDevuelveError(t *testing.T) {
	s := New(newTestDB(t))

	if _, err := s.GetByID(999); err == nil {
		t.Error("GetByID de un ID inexistente debe fallar")
	}
}

func TestGetRandomDevuelveUnaFila(t *testing.T) {
	s := New(newTestDB(t))

	if _, err := s.Crear(nuevaFrase("Única", "orig", "Buda", "budismo")); err != nil {
		t.Fatalf("Crear: %v", err)
	}

	f, err := s.GetRandom()
	if err != nil {
		t.Fatalf("GetRandom: %v", err)
	}
	if f.ID == 0 || f.Frase == "" {
		t.Errorf("esperaba una fila válida, obtuve %+v", f)
	}
}

func TestGetRandomTablaVaciaDevuelveError(t *testing.T) {
	s := New(newTestDB(t))

	if _, err := s.GetRandom(); err == nil {
		t.Error("GetRandom sobre tabla vacía debe fallar")
	}
}

func TestUpdatePersisteCambios(t *testing.T) {
	s := New(newTestDB(t))

	creada, err := s.Crear(nuevaFrase("Vieja", "orig", "Autor", "cat"))
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	actualizada := nuevaFrase("Nueva", "orig2", "Autor2", "cat2")
	if _, err := s.Update(creada.ID, actualizada); err != nil {
		t.Fatalf("Update: %v", err)
	}

	obtenida, err := s.GetByID(creada.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if obtenida.Frase != "Nueva" || obtenida.FraseIdiomaOriginal != "orig2" ||
		obtenida.Autor != "Autor2" || obtenida.Categoria != "cat2" {
		t.Errorf("campos no persistidos: %+v", obtenida)
	}
}

func TestDeleteEliminaLaFila(t *testing.T) {
	s := New(newTestDB(t))

	creada, err := s.Crear(nuevaFrase("Borrar", "orig", "Autor", "cat"))
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}

	if err := s.Delete(creada.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.GetByID(creada.ID); err == nil {
		t.Error("la fila eliminada no debe encontrarse")
	}
}
