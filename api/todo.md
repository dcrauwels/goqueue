# Must-have
- [ ] Implement refresh token rotation in handler_auth.go
- [ ] Update docs/api.md
- [ ] Replace the frontend cookie handling with a backend based system using HTTP-Only cookies instead
- [ ] 401 status codes when access token is expired
- [ ] those 401 status codes being caught and leading to a POST request to /api/refresh
- [ ] and those 401 status codes being retried with a new access token

## Cookie authentication implementation
- [ ] Unify the access token expiration timer through an environment variable stored in cfg (tough)
- [ ] Same for the refresh token expiration timer
- [ ] The logic in the AuthMiddleware function is very ugly right now (if if if)
- [ ] Currently the main hinge is http.ErrNoCookie but this should also trigger if the access token is expired, right?

# Nice to have