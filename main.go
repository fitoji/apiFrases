package main

import (
	"aprende-golang/internal/service"
	"aprende-golang/internal/store"
	"aprende-golang/internal/transport"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	//conectar a SQLite
	db, err := sql.Open("sqlite3", "./frases.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//crear el table si no existe.
	q := `
			CREATE TABLE IF NOT EXISTS frases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			frase TEXT NOT NULL,
			original TEXT NOT NULL,
			autor TEXT NOT NULL,
			categoria TEXT NOT NULL
		)
	`
	if _, err := db.Exec(q); err != nil {
		log.Fatal(err.Error())
	}

	//inyectar nuestras dependencias.
	fraseStore := store.New(db)
	fraseService := service.New(fraseStore)
	fraseHandler := transport.New(fraseService)

	// configurar las rutas
	http.HandleFunc("/frases", fraseHandler.HandleFrases)
	http.HandleFunc("/frases/", fraseHandler.HandleFrasePorID)
	http.HandleFunc("/frases/random", fraseHandler.HandleFraseRandom)

	dbPath, err := filepath.Abs("./frases.db")
	if err == nil {
		fmt.Println("💾 Base de datos en:", dbPath)
	}

	fmt.Println("🚀 Servidor ejecutandose en http://localhost:8000")
	fmt.Println("📡 API Endpoints:")
	fmt.Println("📂 GET /frases         -Obtener todas las frases")
	fmt.Println("📂 GET /frases/random         -Obtener una frase random")
	fmt.Println("📝 POST /frases        -Crear una nueva frase")
	fmt.Println("🔍 GET /frases/{id}    -Obtener un libro especifico")
	fmt.Println("✏️  PUT /frases/{id}    -Actualizar una frase")
	fmt.Println("🗑️  DELETE /frases/{id} -Eliminar una frase")

	//empezar y escuchar al servidor
	log.Fatal(http.ListenAndServe(":8000", nil))
}
