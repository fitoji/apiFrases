package store

import (
	"aprende-golang/internal/model"
	"database/sql"
	"errors"
	"fmt"
)

type Store interface {
	GetAll() ([]*model.Frase, error)
	GetByID(id int) (*model.Frase, error)
	GetRandom() (*model.Frase, error)
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

	frases := make([]*model.Frase, 0)
	for rows.Next() {
		b := &model.Frase{}
		if err := rows.Scan(&b.ID, &b.Frase, &b.FraseIdiomaOriginal, &b.Autor, &b.Categoria); err != nil {
			return nil, err
		}
		frases = append(frases, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return frases, nil
}

func (s *store) GetByID(id int) (*model.Frase, error) {
	q := `SELECT id,frase,original,autor,categoria FROM frases WHERE id=$1`

	f := &model.Frase{}
	err := s.db.QueryRow(q, id).Scan(&f.ID, &f.Frase, &f.FraseIdiomaOriginal, &f.Autor, &f.Categoria)
	if err != nil {
		return nil, err
	}
	return f, nil
}
func (s *store) GetRandom() (*model.Frase, error) {
	q := `SELECT id,frase,original,autor,categoria FROM frases ORDER BY RANDOM() LIMIT 1`

	f := &model.Frase{}
	err := s.db.QueryRow(q).Scan(&f.ID, &f.Frase, &f.FraseIdiomaOriginal, &f.Autor, &f.Categoria)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("no hay frases en la base: %w", err)
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *store) Crear(frase *model.Frase) (*model.Frase, error) {
	q := `INSERT INTO frases (frase, original, autor, categoria) VALUES ($1,$2,$3,$4) RETURNING id`

	if err := s.db.QueryRow(q, frase.Frase, frase.FraseIdiomaOriginal, frase.Autor, frase.Categoria).Scan(&frase.ID); err != nil {
		return nil, err
	}
	return frase, nil
}

func (s *store) Update(id int, frase *model.Frase) (*model.Frase, error) {
	q := `UPDATE frases SET frase=$1, original=$2, autor=$3, categoria=$4 WHERE id=$5`

	res, err := s.db.Exec(q, frase.Frase, frase.FraseIdiomaOriginal, frase.Autor, frase.Categoria, id)
	if err != nil {
		return nil, err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if aff == 0 {
		return nil, sql.ErrNoRows
	}
	frase.ID = id
	return frase, nil
}
func (s *store) Delete(id int) error {
	q := `DELETE FROM frases WHERE id=$1`

	_, err := s.db.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}
