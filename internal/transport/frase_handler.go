package transport

import (
	"aprende-golang/internal/model"
	"aprende-golang/internal/service"
	"encoding/json"
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(frases)

	case http.MethodPost:
		var frase model.Frase
		if err := json.NewDecoder(r.Body).Decode(&frase); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		creado, err := h.service.CrearFrase(frase)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(creado)

	default:
		http.Error(w, "Metodo no disponible!", http.StatusMethodNotAllowed)
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
			http.Error(w, "No lo encontramos!", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(frase)
	case http.MethodPut:
		var frase model.Frase
		if err := json.NewDecoder(r.Body).Decode(&frase); err != nil {
			http.Error(w, "input invalido", http.StatusBadRequest)
			return
		}

		updated, err := h.service.ActualizarFrase(id, frase)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)

	case http.MethodDelete:
		if err := h.service.Borrar(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "metodo no disponible", http.StatusMethodNotAllowed)

	}
}
