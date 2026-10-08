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

## Autenticación

Las lecturas (`GET`) son públicas. Las escrituras (`POST`, `PUT`, `DELETE`) exigen la cabecera `X-API-Key` con el valor de la variable de entorno `API_KEY`:

- `API_KEY` es **opcional**: si no está definida, las escrituras responden `401` (fail-closed: una API sin clave configurada no acepta escrituras).
- Con `API_KEY` definida, cualquier método distinto de `GET`/`OPTIONS` sin la cabecera correcta responde `401 credenciales invalidas`.
- El preflight `OPTIONS` lo intercepta el middleware CORS antes que el de autenticación; los navegadores no necesitan la cabecera.

```bash
API_KEY=MI_CLAVE_SECRETA go run .

curl -X POST http://localhost:8000/frases \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: MI_CLAVE_SECRETA' \
  -d '{"frase":"El dolor es inevitable, el sufrimiento es opcional.","original":"Sabbe saṅkhārā aniccā","autor":"Buda","categoria":"budismo"}'
# → 201 + la frase creada; sin la cabecera o con clave incorrecta → 401 "credenciales invalidas"
```

## Códigos de estado

| Código | Cuándo ocurre |
|--------|---------------|
| `200` / `201` / `204` | Éxito (consulta / creación / borrado). |
| `400` | JSON mal formado, o `{id}` no es numérico. |
| `401` | Escritura sin `X-API-Key` válida, o `API_KEY` sin definir. |
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

La API funciona en dos modos:

- **Sin `DATABASE_URL`** → SQLite local en `./frases.db` (desarrollo local, sin cambios: `go run .`).
- **Con `DATABASE_URL`** → PostgreSQL (usado en Render). El driver, el DSN y el DDL cambian; las queries son idénticas en ambos motores.

Tabla `frases` creada automáticamente al arrancar (modo SQLite):

```sql
CREATE TABLE IF NOT EXISTS frases (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  frase TEXT NOT NULL,
  original TEXT NOT NULL,
  autor TEXT NOT NULL,
  categoria TEXT NOT NULL
);
```

En modo Postgres el `id` usa `INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY` en lugar de `AUTOINCREMENT`.

### Deploy en Render (Postgres free tier)

1. En el dashboard de Render creá una instancia de PostgreSQL **Free**.
2. Conectala a tu web service: Render inyecta `DATABASE_URL` automáticamente.
3. Deploy. Al arrancar la tabla se crea sola.
4. Re-sembrar los datos (la base local no viaja a Render):

```bash
API_URL=https://<tu-servicio>.onrender.com/frases python3 seed_frases.py
```

`seed_frases.py` usa, en orden: variable `API_URL` → primer argumento de la CLI → `http://localhost:8000/frases`.

⚠️ **Caveat del free tier**: Render almacena 1 GB y la instancia **vence a los 30 días de su creación** (hay un período de gracia de 14 días antes de borrar los datos; se permite 1 instancia free por workspace). Ver [render.com/docs/free](https://render.com/docs/free).

## Límites actuales

- CORS incluido: por defecto responde `Access-Control-Allow-Origin: *` y maneja los preflights `OPTIONS`. Para restringir orígenes: `CORS_ALLOWED_ORIGINS="https://app.com,https://admin.com" go run .` (solo esos orígenes reciben las cabeceras).
- Tests de store (`internal/store`) y de CORS (`internal/transport`), en memoria con SQLite.
- Autenticación por API key: las lecturas (`GET`) son públicas y las escrituras requieren `X-API-Key` cuando `API_KEY` está definida (ver [Autenticación](#autenticación)).
- Puerto fijo en `8000`.

## Documentación relacionada

- `Notas.md`, `go01.md`, `goRoutine.md` — apuntes de aprendizaje de Go.
