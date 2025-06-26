# /api/users
## POST
HTTP request with following parameters in JSON:
- email: user email address. Unique.
- password: user password.
- full_name: first and last name for user.
## PUT
# /api/login
## POST
# /api/refresh
## POST
# /api/logout
## POST
# /api/visitors
## POST
## PUT (/api/visitors/{visitor_id})
## GET
Two query parameters:
- purpose: takes any purpose name, filters by specified purpose. Purposes are user defined.
- status: takes integer {0-9?}.

