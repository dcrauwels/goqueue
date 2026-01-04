# goqueue
Queue management backend in Go

# requirements
## PostgreSQL
- Install PostgreSQL as a running service.
- Log in (if default address and port): `psql "postgres://postgres:postgres@localhost:5432"`
- Create goqueue database: `CREATE DATABASE goqueue;`
- Set user password.

## Goose
Run up migration from the sql/schema directory.

## .env
A .env file must manually be created. Please register the following variables:
- DB_URL: postgres connection string with ?sslmode=disable query marameter. E.g.: `DB_URL = "postgres://postgres:postgres@localhost:5432/goqueue"`
- ENV: required to be set to "dev" to access certain /admin endpoints.
- SECRET: used for generating JWTs.
- ACCESSTOKENDURATION: access token expiration time (in minutes).
- REFRESHTOKENDURATION: refresh token expiration time (in weeks).
- PUBLICIDLENGTH: length (in characters) of public-facing IDs for all database entries. Note that this applies to both API calls and urls.
- RESETTIME: timestamp in 24h notation for when all visitors and servicelogs are set to be completed or inactive, respectively. Used to clean up visitors that were not properly handled by operators between working days. Leave empty to disable this feature.

## dependencies
- go get github.com/jaevor/go-nanoid

# usage
## endpoints
Refer to ./docs/api.md for information about endpoints and their usage.



