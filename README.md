# goqueue
Queue management backend in Go

# requirements
## PostgreSQL
- Install PostgreSQL
- Log in and `CREATE DATABASE goqueue;`
- Set user password.

## Goose
Run up migration from the sql/schema directory.

# usage
## endpoints
user management
POST /api/users takes JSON with fields `email` and `password` 
PUT /api/users takes JSON with fields `email` and `password`

authentication    
POST /api/login
POST /api/refresh
POST /api/logout


