package store

import (
	"aprende-golang/internal/model"
	"database/sql"
)

type Store interface {
	GetAll() ([]*model.Frase, error)
	GetByID(id int) (*model.Frase, error)
	Crear(frase *model.Frase) (*model.Frase, error)
	Update(id int, frase *model.Frase) (*model.Frase, error)
	Delete(id int) error
}
type store struct {
	db *sql.DB
}

func New(db *sql.DB) Store {
	return &store{db: db}
}

func (s *store) GetAll() ([]*model.Frase, error) {
	q := `SELECT id,frase,original,autor,categoria FROM frases`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var frases []*model.Frase
	for rows.Next() {
		b := &model.Frase{}
		if err := rows.Scan(&b.ID, &b.Frase, &b.FraseIdiomaOriginal, &b.Autor, &b.Categoria); err != nil {
			return nil, err
		}
		frases = append(frases, b)
	}
	return frases, nil
}

func (s *store) GetByID(id int) (*model.Frase, error) {
	q := `SELECT id,frase,original,autor,categoria FROM frases WHERE id=?`

	f := &model.Frase{}
	err := s.db.QueryRow(q, id).Scan(&f.ID, &f.Frase, &f.FraseIdiomaOriginal, &f.Autor, &f.Categoria)
	if err != nil {
		return nil, err
	}
	return f, nil
}
func (s *store) Crear(frase *model.Frase) (*model.Frase, error) {
	q := `INSERT INTO frases (frase, original, autor, categoria) VALUES (?,?,?,?)`

	resp, err := s.db.Exec(q, frase.Frase, frase.FraseIdiomaOriginal, frase.Autor, frase.Categoria)
	if err != nil {
		return nil, err
	}
	id, err := resp.LastInsertId()
	if err != nil {
		return nil, err
	}
	frase.ID = int(id)
	return frase, nil
}

func (s *store) Update(id int, frase *model.Frase) (*model.Frase, error) {
	q := `UPDATE frases SET frase=?, original=?, autor=?, categoria=? WHERE id= ?`

	_, err := s.db.Exec(q, frase.Frase, frase.FraseIdiomaOriginal, frase.Autor, frase.Categoria, id)
	if err != nil {
		return nil, err
	}
	frase.ID = id
	return frase, nil
}
func (s *store) Delete(id int) error {
	q := `DELETE from frases WHERE id =?`

	_, err := s.db.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}
