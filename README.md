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
- ENV: "dev" for certain /admin endpoints
- ACCESS_TOKEN_DURATION: access token expiration time in minutes

# usage
## endpoints
user management
POST /api/users takes JSON with fields `email` and `password` 
PUT /api/users takes JSON with fields `email` and `password`

authentication    
POST /api/login
POST /api/refresh
POST /api/logout


