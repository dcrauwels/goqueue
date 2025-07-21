# /api/users

Endpoint for users, which represent the employees calling visitors to their desks. Users have accounts that are static in time and authenticate themselves with both an access and a refresh token.

**Response parameters for all requests to /api/users:**

- `id`: UUID, unique, not nullable. Key value identifying user in database.
- `created_at`: timestamp, not nullable. Describes the moment in time the user account was created.
- `updated_at`: timestamp, not nullable. Describes the last time the user account database row was updated.
- `email`: string, unique, not nullable. Describes user email address.
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.
- `is_admin`: boolean, not nullable. Describes whether the user has admin status. True means the user has admin status.
- `is_active`: boolean, not nullable. Describes whether the user account is active. True means the account is active. Accounts set to false will be rejected at the /api/login endpoint and by user authentication middleware when trying to access other authentication required endpoints.

## POST

Request to create a new user. Always creates a non-admin user. Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint. Requires the accessing user to havse is_admin status.

**Request parameters for POST /api/users:**

- `email`: string, unique, not nullable. Describes user email address.
- `password`: string, not nullable. Describes user password.
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.

**Response parameters for POST /api/users:**

See above.

## PUT

Can be sent both to the generic /api/users endpoint and to a specific user UUID at /api/users/{user_id}.

### PUT /api/users

Meant for altering the active user's own account.Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint.

**Request parameters for PUT /api/users:**

- `email`: string, unique, not nullable. Describes user email address. 
- `password`: string, not nullable. Describes user password.
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.

**Response parameters for PUT /api/users:**

See above.


### PUT /api/users/{user_id}

Meant for altering any user's account. This requires admin status. Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint. Requires the user to have is_admin status.

**Request parameters for PUT /api/users/{user_id}:**

- email: string, unique, not nullable. Describes user email address. 
- full_name: string, nullable. Describes first, possibly middle and last name for user.
- is_admin: boolean, not nullable. Describes whether the user has admin status. True means the user has admin status.
- is_active: boolean, not nullable. Describes whether the user account is active. True means the account is active. Accounts set to false will be rejected at the /api/login endpoint and by user authentication middleware when trying to access other authentication required endpoints.

**Response parameters for PUT /api/users/{user_id}:**

See above.

# /api/login

Endpoint for logging in to a user account. Visitors do not need to login as they do not have refresh tokens. 

**Response parameters for all requests to /api/login:**

- `id`: UUID, unique, not nullable. Key value identifying user in database.
- `created_at`: timestamp, not nullable. Describes the moment in time the user account was created.
- `updated_at`: timestamp, not nullable. Describes the last time the user account database row was updated.
- `email`: string, unique, not nullable. Describes user email address.
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.
- `is_admin`: boolean, not nullable. Describes whether the user has admin status. True means the user has admin status.
- `is_active`: boolean, not nullable. Describes whether the user account is active. True means the account is active. Accounts set to false will be rejected at the /api/login endpoint and by user authentication middleware when trying to access other authentication required endpoints.
- `user_access_token`: string, nullable. Describes a JSON Web Token (JWT) that authenticates the current user. Always paired with a refresh token. Note that the access token is stateless and is not stored on the server. Encoded with `bcrypt`. The lifespan of this token is defined together with that of the corresponding cookie in the .env variable. (See readme.md.)
- `user_refresh_token`: string, nullable. Describes a refresh token that is stored on the server in database. Is a hexadecimal encoding of 32 bytes of randomly generated data. The lifespan of this token is defined together with that of the corresponding cookie in the .env variable. (See readme.md.)

## POST

Request to get granted two cookies: "user_access_token" and "user_refresh_token". The lifespans for these cookies are defined in the .env variable. Refer to the readme.md file for more information on setting this up.

**Request parameters:**

- `email`: string, unique, not nullable. Describes user email address.
- `password`: string, not nullable. Describes user password.

**Response parameters:**

See above.

# /api/refresh

## POST

**Request parameters:**

- 

**Response parameters:**

- 

# /api/logout

## POST

**Request parameters:**

- 

**Response parameters:**

- 

# /api/visitors

## POST

**Request parameters:**

- 

**Response parameters:**

- 

## PUT

Can only be sent to an endpoint with a specific visitor UUID.

### PUT /api/visitors/{visitor_id}

**Request parameters:**

- 

**Response parameters:**

- 

## GET

**Request parameters:**

- 

**Response parameters:**

- 

