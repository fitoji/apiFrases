package service

import (
	"aprende-golang/internal/model"
	"aprende-golang/internal/store"
	"errors"
)

type Service struct {
	store store.Store
}

func New(s store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) ObtenTodasLasFrases() ([]*model.Frase, error) {
	return s.store.GetAll()
}

func (s *Service) ObtenFrasePorID(id int) (*model.Frase, error) {
	return s.store.GetByID(id)
}

func (s *Service) ObtenFraseRandom() (*model.Frase, error) {
	return s.store.GetRandom()
}

func (s *Service) CrearFrase(frase model.Frase) (*model.Frase, error) {
	if frase.Frase == "" {
		return nil, errors.New("necesitamos una frase")
	}
	return s.store.Crear(&frase)
}

func (s *Service) ActualizarFrase(id int, frase model.Frase) (*model.Frase, error) {
	if frase.Frase == "" {
		return nil, errors.New("necesitamos una frase")
	}
	return s.store.Update(id, &frase)
}

func (s *Service) Borrar(id int) error {
	return s.store.Delete(id)
}
