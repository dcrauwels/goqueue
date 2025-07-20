# /api/users
Endpoint for users, which represent the employees calling visitors to their desks. Users have accounts that are static in time and authenticate themselves with both an access and a refresh token.
## POST
Request to create a new user. Always creates a non-admin user.
**Request parameters:**
- email: string, unique, not nullable. Describes user email address.
- password: string, not nullable. Describes user password.
- full_name: string, nullable. Describes first, possibly middle and last name for user.
**Response parameters:**
- 
Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint. Requires the user to havse is_admin status.
## PUT
Can be sent both to the generic /api/users endpoint and to a specific user UUID at /api/users/{user_id}.
### PUT /api/users
Meant for altering the active user's own account.
**Request parameters:**
- email: string, unique, not nullable. Describes user email address. 
- password: string, not nullable. Describes user password.
- full_name: string, nullable. Describes first, possibly middle and last name for user.
**Response parameters:**
- 
Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint.
### PUT /api/users/{user_id}
Meant for altering any user's account. This requires admin status.
**Request parameters:**
- email: string, unique, not nullable. Describes user email address. 
- full_name: string, nullable. Describes first, possibly middle and last name for user.
- is_admin: boolean, not nullable. Describes whether the user has admin status. True means the user has admin status.
- is_active: boolean, not nullable. Describes whether the user account is active. True means the account is active. Accounts set to false will be rejected at the /api/login endpoint and by user authentication middleware when trying to access other authentication required endpoints.
**Response parameters:**
- 
Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint. Requires the user to havse is_admin status.


# /api/login
Endpoint for logging in to a user account. Visitors do not need to login as they do not have refresh tokens. 
## POST
Request to get granted two cookies: "user_access_token" and "user_refresh_token". The lifespans for these cookies are defined in the .env variable. Refer to the readme.md file for more information on setting this up.
**Response parameters:**
- 
**Response parameters:**
- 
# /api/refresh
## POST
**Response parameters:**
- 
**Response parameters:**
- 
# /api/logout
## POST
**Response parameters:**
- 
**Response parameters:**
- 
# /api/visitors
## POST
**Response parameters:**
- 
**Response parameters:**
- 
## PUT
Can only be sent to an endpoint with a specific visitor UUID.
### PUT /api/visitors/{visitor_id}
**Response parameters:**
- 
**Response parameters:**
- 
## GET
**Response parameters:**
- 
**Response parameters:**
- 


