# Must-have
- [ ] Implement refresh token rotation in handler_auth.go
- [ ] Update docs/api.md
- [ ] Replace the frontend cookie handling with a backend based system using HTTP-Only cookies instead
- [ ] 401 status codes when access token is expired
- [ ] those 401 status codes being caught and leading to a POST request to /api/refresh
- [ ] and those 401 status codes being retried with a new access token
- [ ] In general /api/refresh needs a fresh coat of paint.
- [ ] Carry over visitor authentication to the cookie structure as well (cry)

## Cookie authentication implementation
- [ ] Unify the access token expiration timer through an environment variable stored in cfg (tough)
- [ ] Same for the refresh token expiration timer
- [x] The logic in the AuthMiddleware function is very ugly right now (if if if)
- [x] Currently the main hinge is http.ErrNoCookie but this should also trigger if the access token is expired, right? (See nice to have todo)

# Nice to have
- [ ] Specify the different errors auth.ValidateJWT can spit out to match the reasons for throwing an error. (Token expired, invalid, etc.)
- [ ] Decide on whether to keep PUT /api/users as well as PUT /api/users/{user_id} or delete the former.

# Other
- [ ] api.HandlerRevokeRefreshToken is NYI (in api/handler_auth.go) 
- [ ] and api.HandlerRevokeRefreshToken also needs an endpoint.
- [ ] api.HandlerGetPurposesByID is NYI (in api/handler_purposes.go)
- [ ] Think about whether api.HandlerGetUsersByID needs authentication or not. Leaning towards yes. Also depends on how I will implement a visitor seeing they've been called.
