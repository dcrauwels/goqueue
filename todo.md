# Must-have
- [x] Implement refresh token rotation in handler_auth.go
- [x] Update docs/api.md
- [x] Replace the frontend cookie handling with a backend based system using HTTP-Only cookies instead
- [x] 401 status codes when access token is expired
- [x] those 401 status codes being caught and leading to a POST request to /api/refresh
- [x] and those 401 status codes being retried with a new access token
- [x] In general /api/refresh needs a fresh coat of paint.
- [x] Carry over visitor authentication to the cookie structure as well (cry)
- [x] api.HandlerRevokeRefreshToken is NYI (in api/handler_auth.go)
- [x] and api.HandlerRevokeRefreshToken also needs an endpoint.
- [x] api.HandlerGetPurposesByID is NYI (in api/handler_purposes.go)
- [x] Think about whether api.HandlerGetUsersByID needs authentication or not. Leaning towards yes. Also depends on how I will implement a visitor seeing they've been called. > currently implemented a check for user auth
- [ ] Stop rotating refresh tokens so much, instead only rotate it when you use it for its purpose of generating an access token? Or is the current approach fine?
- [ ] What is auth.VisitorsByID supposed to do? (in auth/auth.go)
- [ ] All of my http.Redirects are wrong. They more or less all point to "/api/login" which is wrong. It should be an HTML login page like /login. (I think.)
- [ ] Currently there are no checks for user.IsActive. This needs to either go in AuthUserMiddleware or in all of the individual user authentication checks in handlers. The bottom line is: do we want to allow a user to present an access / refresh token for an inactive account and get that ID added to their context? > No, we don't, so it should be blocked at the AuthUserMiddleware level, where we clear the cookie, throw a 401 Unauthorized error, clear cookies and send them to login. (Also see previous todo.)
- [ ] What range of statuses will be allowed? There are multiple NYI's for this, mostly in auth_visitors.go.

## Cookie authentication implementation
- [x] Unify the access token expiration timer through an environment variable stored in cfg (tough)
- [x] Same for the refresh token expiration timer
- [x] The logic in the AuthMiddleware function is very ugly right now (if if if) > no wrong
- [x] Currently the main hinge is http.ErrNoCookie but this should also trigger if the access token is expired, right? (See nice to have todo)
- [x] Bookmark at line 120 ish of makeAuthMiddleware function: what to do when the client provides an access token and refresh token cookie, but the refresh token itself is invalid (while the access token *is* valid)? > you reset both cookies and redirect to the login page. This applies to all unexpected states.
- [ ] Sanity check shower thought: say you send a request directly to let's say POST /api/users (which requires user auth and is_admin status) without a cookie that defines user authentication, but with a custom "user_id" context key with correct value for a user with admin status. Will that pass? Logically speaking: the AuthUserMiddleware function will see no access and refresh token cookies are sent with the request and just pass the whole thing to the handler as is. > The solution is to make AuthUserMiddleware pass a null user_id value to the handler if no valid cookie is presented. Though this make introduce obscure authentication bugs.
- [x] Make sure that in api.HandlerLoginUser (in handler_auth.go) that if a POST request is sent to /api/login while an active refresh token is available for this user_id, that it is revoked. There is an edge case where this is possible. > should be redirected to /api/refresh instead
- [ ] Shower thought: is visitor authentication through cookies even necessary? Currently, the only place it is used, is to send GET requests to /api/visitors/{visitor_id}. But is it really necessary? The alternative method is to simply give visitors the URI to their visitor status page and go from there. 

## Service log implementation
- [ ] Define endpoints for /api/servicelogs. Probably POST, GET, PUT.
- [ ] Write handlers for all of the aforementioned endpoints.
- [ ] We have the same authentication issue for visitors that we have for GET /api/visitors/{visitor_id}. Basically the question is: if a third party that isn't the visitor themselves knows the URI to the visitor status page and can get information from the service log, is that a problem? Does it matter if someone else can see visitors being called?
- [ ] Why is it necessary again to have both a visitor and a service log implementation? Given that a visitor only goes in one direction: from waiting, to serving, to served, what does the log add?

# Nice to have
- [x] Specify the different errors auth.ValidateJWT can spit out to match the reasons for throwing an error. (Token expired, invalid, etc.) > turns out the JWT package has these predefined.
- [ ] Decide on whether to keep PUT /api/users as well as PUT /api/users/{user_id} or delete the former.

# Other
