# API de Frases

Backend REST en Go + SQLite para administrar y consultar frases: CRUD completo más un endpoint de frase aleatoria. Sirve en el puerto `8000` y guarda todo en `./frases.db`.

## Inicio rápido

1. Requisitos: Go instalado (la dependencia `mattn/go-sqlite3` compila con cgo).
2. Levantar el servidor: `go run .`
3. Verificar: `curl http://localhost:8000/frases/random` → debe responder `200` con el JSON de una frase.

Para cargar 10 frases de ejemplo (con el servidor corriendo): `python3 seed_frases.py`

## Endpoints

| Método | Ruta | Descripción | Éxito |
|--------|------|-------------|-------|
| `GET` | `/frases` | Todas las frases | `200` |
| `POST` | `/frases` | Crear una frase | `201` |
| `GET` | `/frases/random` | Una frase al azar | `200` |
| `GET` | `/frases/{id}` | Frase por ID | `200` |
| `PUT` | `/frases/{id}` | Actualizar una frase | `200` |
| `DELETE` | `/frases/{id}` | Eliminar una frase | `204` |

Cualquier método no listado para una ruta responde `405`.

## Esquema de una frase

```json
{
  "id": 1,
  "frase": "No hay camino hacia la felicidad, la felicidad es el camino.",
  "original": "Natthi santaraṁ sukhaṁ",
  "autor": "Buda",
  "categoria": "budismo"
}
```

| Campo | Tipo | Obligatorio | Descripción |
|-------|------|-------------|-------------|
| `id` | number | Solo salida | Autogenerado por SQLite; se ignora en los requests. |
| `frase` | string | Sí | Texto de la frase. Único campo validado por el servicio. |
| `original` | string | Sí (nivel DB) | Texto en el idioma original. |
| `autor` | string | Sí (nivel DB) | Autor de la frase. |
| `categoria` | string | Sí (nivel DB) | Categoría para agrupar (p. ej. `budismo`). |

## Ejemplos

### Crear

```bash
curl -X POST http://localhost:8000/frases \
  -H 'Content-Type: application/json' \
  -d '{"frase":"El dolor es inevitable, el sufrimiento es opcional.","original":"Sabbe saṅkhārā aniccā","autor":"Buda","categoria":"budismo"}'
# → 201 + la frase creada con su id asignado
```

### Frase aleatoria

```bash
curl http://localhost:8000/frases/random
# → 200 + una frase al azar; si la tabla está vacía → 404 "no hay frases en la base"
```

### Consultar / actualizar / eliminar

```bash
curl http://localhost:8000/frases/1        # → 200 la frase, o 404 si no existe
curl -X PUT http://localhost:8000/frases/1 \
  -H 'Content-Type: application/json' \
  -d '{"frase":"Texto nuevo","original":"Original nuevo","autor":"Autor","categoria":"cat"}'
curl -X DELETE http://localhost:8000/frases/1   # → 204 sin cuerpo
```

## Códigos de estado

| Código | Cuándo ocurre |
|--------|---------------|
| `200` / `201` / `204` | Éxito (consulta / creación / borrado). |
| `400` | JSON mal formado, o `{id}` no es numérico. |
| `404` | ID inexistente, o `/frases/random` sobre tabla vacía. |
| `405` | Método HTTP no soportado en la ruta. |
| `500` | Error de base de datos, `frase` vacía (`necesitamos una frase`), u otro campo obligatorio faltante (restricción `NOT NULL`). |

## Casos borde conocidos

- `GET /frases` sin filas devuelve `null`, no `[]`.
- `PUT /frases/{id}` con un ID inexistente responde `200` (no verifica que la fila exista).
- `DELETE /frases/{id}` con un ID inexistente responde `204`.
- Los errores responden texto plano (`Content-Type: text/plain`), no JSON.

## Arquitectura

| Capa | Paquete | Rol |
|------|---------|-----|
| Modelo | `internal/model` | Entidad `Frase` y sus tags JSON. |
| Persistencia | `internal/store` | Interface `Store` + SQL sobre SQLite. |
| Servicio | `internal/service` | Reglas de negocio (valida `frase` no vacía). |
| Transporte | `internal/transport` | Handlers HTTP y serialización JSON. |
| Composición | `main.go` | Inyección de dependencias, DDL y rutas. |

## Base de datos

Tabla `frases` creada automáticamente al arrancar:

```sql
CREATE TABLE IF NOT EXISTS frases (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  frase TEXT NOT NULL,
  original TEXT NOT NULL,
  autor TEXT NOT NULL,
  categoria TEXT NOT NULL
);
```

## Límites actuales

- Sin CORS: un frontend en otro origen necesita un proxy o cabeceras CORS.
- Sin tests automatizados.
- Sin autenticación: el CRUD está abierto.
- Puerto fijo en `8000`.

## Documentación relacionada

- `Notas.md`, `go01.md`, `goRoutine.md` — apuntes de aprendizaje de Go.
