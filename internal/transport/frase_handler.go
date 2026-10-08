package transport

import (
	"aprende-golang/internal/model"
	"aprende-golang/internal/service"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type FraseHandler struct {
	service *service.Service
}

func New(s *service.Service) *FraseHandler {
	return &FraseHandler{
		service: s,
	}
}

func (h *FraseHandler) HandleFrases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		frases, err := h.service.ObtenTodasLasFrases()
		if err != nil {
			log.Printf("error al obtener frases: %v", err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(frases); err != nil {
			log.Printf("error al escribir respuesta de frases: %v", err)
		}

	case http.MethodPost:
		var frase model.Frase
		if err := json.NewDecoder(r.Body).Decode(&frase); err != nil {
			log.Printf("error al decodificar frase: %v", err)
			http.Error(w, "input invalido", http.StatusBadRequest)
			return
		}
		creado, err := h.service.CrearFrase(frase)
		if err != nil {
			if errors.Is(err, service.ErrFraseVacia) {
				http.Error(w, "necesitamos una frase", http.StatusBadRequest)
				return
			}
			log.Printf("error al crear frase: %v", err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(creado); err != nil {
			log.Printf("error al escribir respuesta de frase creada: %v", err)
		}

	default:
		http.Error(w, "Metodo no disponible!", http.StatusMethodNotAllowed)
	}
}
func (h *FraseHandler) HandleFraseRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no disponible!", http.StatusMethodNotAllowed)
		return
	}

	frase, err := h.service.ObtenFraseRandom()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "no hay frases en la base", http.StatusNotFound)
			return
		}
		log.Printf("error al obtener frase random: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(frase); err != nil {
		log.Printf("error al escribir respuesta de frase random: %v", err)
	}
}

func (h *FraseHandler) HandleFrasePorID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/frases/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "no lo encontre!", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		frase, err := h.service.ObtenFrasePorID(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No lo encontramos!", http.StatusNotFound)
				return
			}
			log.Printf("error al obtener frase por id: %v", err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(frase); err != nil {
			log.Printf("error al escribir respuesta de frase por id: %v", err)
		}
	case http.MethodPut:
		var frase model.Frase
		if err := json.NewDecoder(r.Body).Decode(&frase); err != nil {
			http.Error(w, "input invalido", http.StatusBadRequest)
			return
		}

		updated, err := h.service.ActualizarFrase(id, frase)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No lo encontramos!", http.StatusNotFound)
				return
			}
			if errors.Is(err, service.ErrFraseVacia) {
				http.Error(w, "necesitamos una frase", http.StatusBadRequest)
				return
			}
			log.Printf("error al actualizar frase: %v", err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(updated); err != nil {
			log.Printf("error al escribir respuesta de frase actualizada: %v", err)
		}

	case http.MethodDelete:
		if err := h.service.Borrar(id); err != nil {
			log.Printf("error al borrar frase: %v", err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "metodo no disponible", http.StatusMethodNotAllowed)

	}
}
