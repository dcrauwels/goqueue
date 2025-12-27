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
- [x] Stop rotating refresh tokens so much, instead only rotate it when you use it for its purpose of generating an access token? Or is the current approach fine?
- [?] What is auth.VisitorsByID supposed to do? (in auth/auth.go)
- [x] All of my http.Redirects are wrong. They more or less all point to "/api/login" which is wrong. It should be an HTML login page like /login. (I think.)
- [x] Currently there are no checks for user.IsActive. This needs to either go in AuthUserMiddleware or in all of the individual user authentication checks in handlers. The bottom line is: do we want to allow a user to present an access / refresh token for an inactive account and get that ID added to their context? > No, we don't, so it should be blocked at the AuthUserMiddleware level, where we clear the cookie, throw a 401 Unauthorized error, clear cookies and send them to login. (Also see previous todo.)
- [ ] What range of statuses will be allowed? There are multiple NYI's for this, mostly in auth_visitors.go.
- [ ] Double-check *all* authentication checks in handlers if user.IsActive is taken into account. Compare to how it's done in HandlerPostDesks in handler_desks.go

## Statuses
- [ ] Think about whether statuses should be hardcoded or user-defined (like purposes)
- [ ] Define statuses, currently implemented as integers, so a map is needed for integers > meaning
- [ ] Implement statuses properly, in the following parts:
- [ ] 1. visitors queries

## GET /api/visitors 
### query parameters
- [ ] At least for the GET /api/visitors endpoint, we have real query parameters. E.g. GET /api/visitors?purpose=fun. I don't think these are implemented properly: they should be optional, in the sense that GET /api/visitors should just return ALL visitors, but I believe it only works right now if both query parameters are called, e.g. GET /api/visitors?purpose=&status=.
- [ ] Filtering by date should be possible.
### HandlerGetVisitorsByPublicID
- [ ] There is a visitor authentication scheme which I believe is deprecated in here. The underlying idea is that only users and the visitor matching the queried PublicID would be authorized to perform this GET request. But I believe the auth.VisitorFromContext() function that is called here no longer works and that the entire visitor authentication scheme has just been abandoned and that I decided it's fine for visitors to just be queried at any point if the PublicID is known.

## Cookie authentication implementation
- [x] Unify the access token expiration timer through an environment variable stored in cfg (tough)
- [x] Same for the refresh token expiration timer
- [x] The logic in the AuthMiddleware function is very ugly right now (if if if) > no wrong
- [x] Currently the main hinge is http.ErrNoCookie but this should also trigger if the access token is expired, right? (See nice to have todo)
- [x] Bookmark at line 120 ish of makeAuthMiddleware function: what to do when the client provides an access token and refresh token cookie, but the refresh token itself is invalid (while the access token *is* valid)? > you reset both cookies and redirect to the login page. This applies to all unexpected states.
- [x] Sanity check shower thought: say you send a request directly to let's say POST /api/users (which requires user auth and is_admin status) without a cookie that defines user authentication, but with a custom "user_id" context key with correct value for a user with admin status. Will that pass? Logically speaking: the AuthUserMiddleware function will see no access and refresh token cookies are sent with the request and just pass the whole thing to the handler as is. > The solution is to make AuthUserMiddleware pass a null user_id value to the handler if no valid cookie is presented. Though this make introduce obscure authentication bugs.
- [x] Make sure that in api.HandlerLoginUser (in handler_auth.go) that if a POST request is sent to /api/login while an active refresh token is available for this user_id, that it is revoked. There is an edge case where this is possible. > should be redirected to /api/refresh instead
- [x] Shower thought: is visitor authentication through cookies even necessary? Currently, the only place it is used, is to send GET requests to /api/visitors/{visitor_id}. But is it really necessary? The alternative method is to simply give visitors the URI to their visitor status page and go from there. > yaba daba this is what I doo

## NanoID implementation
- [x] Think about where the public_id is and isn't relevant. (Frontend vs. backend API.) > both, UUID is only for database robustness
- [x] Migrate the following tables to include a 'public_id' row: users, visitors, desks, service_logs, purposes, refresh_tokens.
- [x] Update the SQL queries to take public_id where relevant. Probably only the CreateX queries.
- [x] Add SQL queries for finding a table row by public_id.
- [x] Update the environment variable handling in main.go by writing a function that uses os.LookupEnv to more robustly handle errors in environment variable setting. 
- [x] Incorporate PUBLICIDLENGTH as an environment variable.
- [x] Add the nanoid package to dependencies in readme.md.
- [?] Update the handlers to invoke the nanoid generator and the pass generated public_id into the updated SQL queries.
- [?] Update the handlers to have the queryparameters and responseparameter structs include public_ids.
- [x] POST & PUT & GET users
- [x] POST & PUT & GET visitors
- [x] POST & PUT desks > need to do the handler_desks.go stuff first
- [x] Refresh tokens will only hold the user public ID as foreign key. This means JWT implementation needs to change.
- [x] Followup on previous todo: fix auth_test.go - specifically the JWT unit test that currently takes a uuid.new() as input.
- [x] POST & PUT refresh_tokens
- [x] POST & PUT purposes
- [x] A number of api paths have '{public_visitor_id}' in them and a number have '{user_public_id}'. Unify to latter style across main.go and handlers.
- [ ] POST & PUT service_logs
- [x] Follow down the road to fix the handler functions.

## handler_desks.go
- [x] Write DB migration to drop the desk number column and instead implement a desk name string. That way you can have desk 'F1' 'S1' etc. There is no real reason to restrict it to numbers other than if you are going to pass the desk number the public ID. But I don't think that makes sense for a number of reasons.
- [x] Update queries for new desks schema.
- [x] POST /api/desks
- [x] PUT /api/desks/{public_desk_id}
- [x] GET /api/desks
- [ ] Should GET /api/desks require authentication? Probably.
- [x] GET /api/desks/{public_desk_id}
- [x] Currently the first 3 functions require userauth, the last doesn't. Is that a problem? > No. Reasoning being that when a servicelog is created for a visitor being called to a desk the desk pid is provided and the visitor needs to know what actual desk (e.g. desk.Name) they are going to. And they can get that information by querying GET /api/desks/{pid_from_service_log}
- [ ] some sort of GET request for active desks, perhaps through a query parameter. Note that I need to resolve the QP issue in /api/visitors first

## Visitor daily_ticket_number implementation
- [x] Write a migration for a ticket_counter table (two columns: date as primary key, last_ticket_number as int)
- [x] Write a migration for the visitors table to take a daily_ticket_number (INT) column
- [x] Write a query to update the ticket_counter table for today
- [x] Update the visitors handlers (should only be for POST /api/visitors)


## Service log implementation
- [x] Define endpoints for /api/servicelogs. Probably POST, GET, PUT.
- [x] Write a migration for the service_logs table to accommodate public ids in tables users, visitors, desks.
- [ ] Define a GET /api/visitors/{visitor_id}/status endpoint. This is meant for a visitor to check their own status ideally.
- [ ] Define a /api/queue endpoint which takes GET requests and is meant for a screen to display all WAITING / CALLED / SERVING visitors.
- [ ] Write handlers for all of the aforementioned endpoints.
- [x] We have the same authentication issue for visitors that we have for GET /api/visitors/{visitor_id}. Basically the question is: if a third party that isn't the visitor themselves knows the URI to the visitor status page and can get information from the service log, is that a problem? Does it matter if someone else can see visitors being called? 
- [x] Why is it necessary again to have both a visitor and a service log implementation? Given that a visitor only goes in one direction: from waiting, to serving, to served, what does the log add?

# Nice to have
- [x] Specify the different errors auth.ValidateJWT can spit out to match the reasons for throwing an error. (Token expired, invalid, etc.) > turns out the JWT package has these predefined.
- [ ] Decide on whether to keep PUT /api/users as well as PUT /api/users/{user_id} or delete the former.
- [ ] Currently GET /api/users requires admin status. Is that actually necessary?
- [ ] There is currently a privacy problem where visitors can be queried historically. The identifying information is really in their name more than anything else. So that needs to be periodically removed from the visitors table, as it's not relevant for statistical purposes either.
- [ ] The scenario where a non-admin user accesses their own user_id under POST /api/revoke should redirect to /api/logout, not just throw a bad request error.

# Other
