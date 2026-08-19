# Redaccion de credenciales en el log de Connect - 2026-08-19

## Archivos modificados

- `mongocc.go` -- `redactMongoURI` y `connectionLogLine` nuevas; la linea de log de `Connect` ya no imprime la URI cruda; `fmt.Errorf(err.Error())` reemplazado por `errors.New(err.Error())` en las dos `CheckMongoError`
- `mongocc_test.go` -- nuevo, primer test del paquete
- `docs/versions/v1.3.1.md` -- nuevo

## Problema

`Connect` imprimia la cadena de conexion completa, con usuario y password:

```go
fmt.Printf("You successfully connected to Mongo: %s db: %s\n", mongoUri, dbName)
```

La linea era **incondicional**. `opts.Debug` recien se lee doce lineas mas abajo,
asi que pasar `Debug: false` no la silenciaba: no habia forma de apagarla desde el
consumidor. Cada llamada a `Connect` escribia las credenciales de esa base a
stdout, y todo lo que captura stdout --logs de contenedor, un agregador de logs,
la salida de CI-- las retenia.

El alcance no es una conexion por servicio. Un servicio con tres bases escribia
tres credenciales distintas en cada arranque.

Es una regresion parcial de v1.3.0: esa version hizo el debug logging configurable
via `ClientOptions.Debug`, pero esta linea quedo afuera del cambio.

## Cambios realizados

### 1. La URI se reconstruye desde scheme y host

```go
func redactMongoURI(mongoUri string) string {
	parsed, err := url.Parse(mongoUri)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return redactedURI
	}
	return parsed.Scheme + "://" + parsed.Host
}
```

| Entrada | Sale al log |
|:--|:--|
| `mongodb://appuser:s3cr3t@cluster.example.com:27017/billing?retryWrites=true` | `mongodb://cluster.example.com:27017` |
| `mongodb+srv://appuser:s3cr3t@cluster.abc.mongodb.net/billing` | `mongodb+srv://cluster.abc.mongodb.net` |
| `mongodb://a.example:27017,b.example:27017/billing?replicaSet=rs0` | `mongodb://a.example:27017,b.example:27017` |
| `://///` | `[redacted]` |

**Por que reconstruir y no borrar el userinfo.** Poner `parsed.User = nil` y
serializar de vuelta es la solucion obvia y **no alcanza**: el userinfo no es el
unico lugar de una cadena de conexion de MongoDB que puede tener un secreto. La
cadena de opciones tambien lleva `tlsCertificateKeyFilePassword` y el token de
sesion de AWS dentro de `authMechanismProperties`. Las dos variantes estan en el
test, y las dos pasan con la version que solo borra el userinfo.

Reconstruir desde scheme y host invierte el default: en vez de sacar lo que hoy
sabemos que es secreto, se conserva unicamente lo que sabemos que no lo es. Una
opcion nueva agregada mañana a la cadena no se filtra por omision.

Lo que se pierde es el path de la URI --normalmente la base de autenticacion. El
valor diagnostico de la linea queda intacto: `dbName` ya se imprime aparte, asi
que "a que servidor me pegue y contra que base" se sigue respondiendo.

### 2. El mensaje se arma en una funcion aparte

`connectionLogLine` existe para que la redaccion sea testeable sin un MongoDB
vivo. Un test que solo cubriera `redactMongoURI` no detectaria una edicion futura
que vuelva a formatear la URI cruda dentro del mensaje, que es exactamente como
las credenciales llegaron ahi.

No es teorico: al mutar el codigo para verificar los tests, reponer la URI cruda
en `Connect` dejo `TestRedactMongoURI` **en verde**. Solo fallo
`TestConnectionLogLineNeverLeaksSecrets`.

### 3. Primer test del paquete

`mongocc_test.go`, 18 casos:

- `TestRedactMongoURI` -- tabla con la salida exacta esperada por forma de URI
- `TestConnectionLogLineNeverLeaksSecrets` -- la asercion de seguridad: un secreto
  centinela en cada lugar donde una cadena de conexion puede tenerlo, y ninguno
  aparece en la linea
- `TestConnectionLogLineKeepsDiagnosticValue` -- guarda el otro lado: redactar la
  linea entera no es una forma aceptable de pasar el test anterior

Los tres se verificaron mutando el codigo de produccion, no solo mirandolos pasar.

### 4. `fmt.Errorf(err.Error())` reemplazado por `errors.New(err.Error())`

No es cosmetico y no es opcional: **el paquete no podia correr ningun test sin
esto.** `go test` corre un subconjunto de `go vet` por default, y el check de
`printf` falla el build con "non-constant format string in call to fmt.Errorf".

Ademas es un bug real: un mensaje de error de Mongo que contenga un `%` se
interpretaba como verbo de formato y salia mangleado (`%!s(MISSING)`).

`errors.New(err.Error())` mantiene el string identico y la misma semantica de
identidad que habia (un error nuevo, sin cadena de wrapping). `return err` seria
mejor --preservaria `errors.Is`-- pero es un cambio de comportamiento observable
y no corresponde meterlo en un fix de seguridad.

## Que NO cambio

- La linea sigue siendo incondicional, no se movio detras de `opts.Debug`. Saber
  a que servidor se pego un servicio al arrancar es util y no es un secreto una
  vez redactada la URI
- `bson.D{{"ping", 1}}` sigue con campos sin key. `go vet ./...` lo reporta, pero
  el subconjunto que corre `go test` no, asi que no bloquea nada
- La interface `MongoFunctions` sigue sin `FindOneAndUpdate`, `InsertMany` ni
  `GetCollection`
- La funcion standalone `CheckMongoError` (deprecated) sigue exportada
- No hay workflow de CI que corra los tests. El repo solo tiene `publish.yml`,
  que dispara en release
