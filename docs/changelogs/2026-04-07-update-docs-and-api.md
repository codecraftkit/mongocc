# Actualizacion de documentacion y API de Connect - 2026-04-07

## Archivos modificados

- `mongocc.go` -- firma de `Connect`, prefijo de logs de debug
- `structs.go` -- struct `ClientOptions` (sin cambios, ya existia)
- `README.md` -- ejemplo de uso actualizado a la API actual
- `LICENSE` -- copyright actualizado a 2025-2026
- `docs/project-review.md` -- nuevo, revision tecnica del proyecto (ingles)
- `docs/project-overview.md` -- nuevo, descripcion general no tecnica (ingles)

## Problema

El README mostraba una API obsoleta (`MongoDataStore`, `Connect` con firma diferente) que no coincidia con el codigo actual. Ademas, `Connect` tenia `Debug` hardcodeado a `true` sin forma de configurarlo.

## Cambios realizados

### 1. Firma de `Connect` acepta `ClientOptions`

```go
// Antes
func Connect(mongoUri string, dbName string) (*MongoQueries, error)
// Debug hardcodeado: Debug: true

// Despues
func Connect(mongoUri string, dbName string, opts *ClientOptions) (*MongoQueries, error)
// Debug configurable via opts.Debug, nil-safe
```

### 2. Prefijo de logs de debug renombrado

| Antes | Despues |
|:------|:--------|
| `[LOG]` | `[Mongocc Log]` |

Aplica a todos los metodos: Find, FindOne, FindOneAndUpdate, InsertOne, InsertMany, UpdateOne, UpdateMany, DeleteOne, DeleteMany, Aggregate, CountDocuments, CheckMongoError.

### 3. README actualizado

- Reemplazo de `MongoDataStore` por `mongocc.Connect()` que retorna `(*MongoQueries, error)`
- Ejemplo usa metodos wrapper (`mq.InsertOne`, `mq.FindOne`) en lugar de operar directo sobre colecciones
- Import de `bson` agregado correctamente
- Key Features actualizadas (debug mode, error classification)
- Version del driver corregida a `mongo-driver/v2`

### 4. LICENSE actualizado

Copyright de `2025` a `2025-2026`.

### 5. Documentacion tecnica y no tecnica generada

Dos archivos nuevos en `docs/`:
- `project-review.md` -- revision tecnica con arquitectura, API, stack, observaciones
- `project-overview.md` -- descripcion orientada a negocio sin jerga tecnica

## Que NO cambio

- La interface `MongoFunctions` sigue sin incluir `FindOneAndUpdate`, `InsertMany` ni `GetCollection`
- La funcion standalone `CheckMongoError` (deprecated) sigue exportada
- No se agregaron tests
