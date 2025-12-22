package model

type Frase struct {
	ID                  int    `json:"id"`
	Frase               string `json:"frase"`
	FraseIdiomaOriginal string `json:"original"`
	Autor               string `json:"autor"`
	Categoria           string `json:"categoria"`
}
