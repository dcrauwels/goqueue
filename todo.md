# Must-have
- [x] Implement refresh token rotation in handler_auth.go
- [ ] Update docs/api.md
- [x] Replace the frontend cookie handling with a backend based system using HTTP-Only cookies instead
- [x] 401 status codes when access token is expired
- [x] those 401 status codes being caught and leading to a POST request to /api/refresh
- [x] and those 401 status codes being retried with a new access token
- [ ] In general /api/refresh needs a fresh coat of paint.
- [x] Carry over visitor authentication to the cookie structure as well (cry)
- [ ] api.HandlerRevokeRefreshToken is NYI (in api/handler_auth.go)
- [ ] and api.HandlerRevokeRefreshToken also needs an endpoint.
- [ ] api.HandlerGetPurposesByID is NYI (in api/handler_purposes.go)
- [ ] Think about whether api.HandlerGetUsersByID needs authentication or not. Leaning towards yes. Also depends on how I will implement a visitor seeing they've been called.

## Cookie authentication implementation
- [x] Unify the access token expiration timer through an environment variable stored in cfg (tough)
- [x] Same for the refresh token expiration timer
- [x] The logic in the AuthMiddleware function is very ugly right now (if if if)
- [x] Currently the main hinge is http.ErrNoCookie but this should also trigger if the access token is expired, right? (See nice to have todo)
- [ ] Bookmark at line 120 ish of makeAuthMiddleware function: what to do when the client provides an access token and refresh token cookie, but the refresh token itself is invalid (while the access token *is* valid)?

# Nice to have
- [x] Specify the different errors auth.ValidateJWT can spit out to match the reasons for throwing an error. (Token expired, invalid, etc.) > turns out the JWT package has these predefined.
- [ ] Decide on whether to keep PUT /api/users as well as PUT /api/users/{user_id} or delete the former.

# Other
- [ ] Make sure that in api.HandlerLoginUser (in handler_auth.go) that if a POST request is sent to /api/login while an active refresh token is available for this user_id, that it is revoked. There is an edge case where this is possible.