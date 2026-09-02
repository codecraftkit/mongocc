# Handles por base sobre un cliente compartido - 2026-09-02

## Archivos modificados

- `mongocc.go` -- `NewFromClient`, `WithDatabase` y `Database` nuevos; `Connect` arma el `MongoQueries` a traves de `NewFromClient`
- `mongocc_handles_test.go` -- nuevo; primeros tests de los constructores
- `README.md` -- seccion "Several databases on one client"
- `docs/versions/v1.4.0.md` -- nuevo

## Problema

`Connect` era la **unica** forma de construir un `MongoQueries`, y cada llamada
abre un `*mongo.Client` nuevo: su propio pool de conexiones, sus propias
goroutines de monitoreo. El campo `db` es privado, asi que un consumidor con un
handle a mano no tenia manera de derivar otro para una segunda base sobre el
mismo cliente.

El caso que lo saco a la luz es multi-tenant: un servicio con **una base por
tenant sobre el mismo deployment**. Su factory llamaba a `Connect` por tenant,
cacheaba cinco minutos, y al vencer volvia a llamar a `Connect` **sin
desconectar el anterior**. Un pool por tenant, y otro mas por tenant cada cinco
minutos. Nada fallaba: los sockets crecian hasta que otra cosa se rompia.

El driver ya tiene la pieza que faltaba: `client.Database(name)` es un handle
sin I/O de red. Lo que no habia era la forma de envolverlo en un `MongoQueries`.

## Cambios realizados

### 1. `NewFromClient` -- constructor sobre un cliente que ya existe

```go
func NewFromClient(client *mongo.Client, dbName string, opts *ClientOptions) *MongoQueries
```

Sin ping y sin linea de log: el cliente ya lo verifico quien lo abrio. `opts`
nil significa `Debug` apagado, igual que en `Connect`. `Connect` ahora termina
llamando a esto, asi que hay una sola forma de armar el struct.

### 2. `WithDatabase` -- otra base, el MISMO cliente

```go
func (mongodb *MongoQueries) WithDatabase(dbName string) *MongoQueries
```

Devuelve un handle nuevo sobre `mongodb.db.Client().Database(dbName)`. Comparte
el pool del receptor, hereda su `Debug`, y no modifica al receptor.

Es la forma multi-tenant: **un cliente por proceso, un handle por tenant.** El
test que lo fija es igualdad de puntero sobre el cliente -- N handles derivados
de una base tienen que reportar el mismo `*mongo.Client`.

### 3. `Database` -- accessor al handle

```go
func (mongodb *MongoQueries) Database() *mongo.Database
```

`Name()` dice contra que base opera; `Client()` es el camino para hacer
`Disconnect` al apagar -- **por el dueño del cliente, una vez**, no por cada
handle derivado. Un `WithDatabase` que desconecte se lleva puestos a todos los
demas.

### 4. Tests sin MongoDB vivo

`mongo.Connect` en driver v2 **no marca**: arranca el monitor de topologia en
background y devuelve. Una URI a un puerto cerrado da un cliente usable cuyos
`Database()` se pueden inspeccionar sin correr una sola operacion. Los cuatro
tests preguntan sobre que cliente esta un handle y que base nombra, y las dos
cosas se responden localmente.

## Que NO cambio

- La firma de `Connect` es la misma; su comportamiento tambien (ping + log)
- La interface `MongoFunctions` sigue sin incluir los metodos nuevos, ni
  `FindOneAndUpdate`, `InsertMany` o `GetCollection`
- La funcion standalone `CheckMongoError` (deprecated) sigue exportada
- `bson.D{{"ping", 1}}` sigue con campos sin key; `go vet` lo reporta y el
  subconjunto de `go test` no
- Sigue sin haber workflow de CI que corra los tests
