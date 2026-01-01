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

## POST /api/users

Request to create a new user. Always creates a non-admin user. Checks for user authentication via "user_access_token" and "user_refresh_token" cookies, which are provided at the /api/login endpoint. Requires the accessing user to havse is_admin status.

**Request parameters for POST /api/users:**

- `email`: string, unique, not nullable. Describes user email address.
- `password`: string, not nullable. Describes user password.
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.

**Response parameters for POST /api/users:**

See above.

## PUT /api/users

Can be sent both to the generic /api/users endpoint and to a specific user UUID at /api/users/{user_id}.

### PUT /api/users (generic)

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

- `email`: string, unique, not nullable. Describes user email address. 
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.
- `is_admin`: boolean, not nullable. Describes whether the user has admin status. True means the user has admin status.
- `is_active`: boolean, not nullable. Describes whether the user account is active. True means the account is active. Accounts set to false will be rejected at the /api/login endpoint and by user authentication middleware when trying to access other authentication required endpoints.

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

## POST /api/login

Request to get granted two cookies: "user_access_token" and "user_refresh_token". The lifespans for these cookies are defined in the .env variable. Refer to the readme.md file for more information on setting this up.

**Request parameters:**

- `email`: string, unique, not nullable. Describes user email address.
- `password`: string, not nullable. Describes user password.

**Response parameters:**

See above.

# /api/refresh

Endpoint for requesting a new access token when the user already has a valid refresh token. Note that both token types are implemented via HTTPOnly cookies.

**Response parameters:**

- 	`token`: string, unique, not nullable. Represents a refresh token.
-   `


## POST /api/refresh

**Request parameters:**

- 

## GET /api/refresh

# /api/logout

Endpoint for deactivating a user's current authentication credentials. This means that clientside, the "user_access_token" and "user_refresh_token" cookies are nulled and serverside, the refresh token is revoked in the database.

**Response parameters for all requests to /api/logout:**
- `id`: UUID, unique, not nullable. Key value identifying user in database.
- `created_at`: timestamp, not nullable. Describes the moment in time the user account was created.
- `updated_at`: timestamp, not nullable. Describes the last time the user account database row was updated.
- `email`: string, unique, not nullable. Describes user email address.
- `full_name`: string, nullable. Describes first, possibly middle and last name for user.
- `is_admin`: boolean, not nullable. Describes whether the user has admin status. True means the user has admin status.
- `is_active`: boolean, not nullable. Describes whether the user account is active. True means the account is active. Accounts set to false will be rejected at the /api/login endpoint and by user authentication middleware when trying to access other authentication required endpoints.
- `user_access_token`: string, null. Describes a JWT access token authenticating the user. Note that the cookie containing this token on the client's end is also nulled.
- `user_refresh_token`: string, null. Note that the cookie containing this token on the client's end is also nulled.

## POST /api/logout

**Request parameters:**

None, the user identity is taken from the HTTP request context. By extension, this means it is taken from the user_access_token cookie and validated against the user_refresh_token cookie.

# /api/visitors
Endpoint for handling visitors, who are models of actual human visitors to the physical location. In terms of permissions, they are placed below users. Users can edit visitors (through PUT /api/visitors) but visitors cannot edit users.

**Response parameters for all requests to /api/visitors:**
- `id`: UUID, unique, not nullable. Key value for identifying visitor in database.
- `created_at`: timestamp, not nullable. Describes the moment in time the visitor entry was created.
- `updated_at`: timestamp, not nullable. Describes the last time the visitor database row was updated.
- `waiting_since`: timestamp, not nullable. Describes the moment in time since when the visitor has been waiting. For purposes of determining which visitor is called (whether FIFO or LIFO).
- `name`: string, nullable. Describes the name of the visitor waiting in line. Note that this is currently nullable while the corresponding field in the POST request is not.
- `purpose_id`: UUID, not nullable. Identifies the visitor chosen purpose in the purpose database.
- `status`: int (32 bit), not nullable. Describes the status of the visitor: waiting, being helped, helped, cancelled by visitor, cancelled by user. NYI.

## POST /api/visitors

**Request parameters:**

- `name`: string, not nullable. Subject to change. Contains the name of the visitor.
- `purpose_id`: UUID, not nullable. Identifies the visitor chosen purpose in the purpose database. There should be a very limited number of purposes ultimately. 

Regarding the notes 'subject to change': I am making the `name` field nullable in a future version. Additionally, I think passing the purpose as a UUID is quite hostile to the user and there is an option to make pass by name instead. The problem is that names are technically not unique values by design.

**Response parameters:**

In addition to the general response parameters described above, a successful POST request also returns:
- `visitor_access_token`: string, not nullable. Describes a JWT access token authenticating the visitor.
Currently, there is no structure for saving the access token as a cookie in the way that this happens for users. In a future version, this will be implemented.

## PUT /api/visitors

Can only be sent to an endpoint with a specific visitor UUID. Meant for altering a visitor's status. Can be performed either by a user or by the visitor in question.

### PUT /api/visitors/{visitor_id}

**Request parameters:**

- `name`: string, not nullable. Subject to change. Contains the name of the visitor.
- `purpose_id`: UUID, not nullable. Identifies the visitor chosen purpose in the purpose database. There should be a very limited number of purposes ultimately. 
- `status`: int (32 bit), not nullable. Describes the status of the visitor: waiting, being helped, helped, cancelled by visitor, cancelled by user. NYI.

**Response parameters:**

See above. Subject to change: there's an argument to be made the visitor access token should be returned as well iff the request is sent with visitor authentication.

## GET /api/visitors

Can be sent both to the generic /api/visitors endpoint and to a specific visitor UUID endpoint. Requests to the generic endpoint will return all visitors and can therefore only be made to users. This is integral to populating a list of visitors to be called. Requests to the specific endpoint can also be made with visitor authentication if the visitor access token matches the UUID where the request is being sent to.

The generic /api/visitors endpoint takes query parameters for GET requests. This can be used to filter for visitors with specific statuses or purposes. The point of this feature is to allow users to generate usable lists of visitors for calling purposes. Example: GET /api/visitors?purpose=finances&status=1

**Query parameters for generic endpoint:**

- `purpose`: public ID referring to the purpose in question. Frontend will need to show the corresponding purpose name for user legibility.
- `status`: integer. NYI. Frontend will need to show the corresponding status name for user legibility.
- `start_date`: ISO 8601 timestamp (YYYY-MM-DD). Inclusive. 
- `end_date`: ISO 8601 timestamp (YYYY-MM-DD). Exclusive. 

**Response parameters:**

Returns either a set of visitors or a single visitor, depending on whether the request is sent to the generic or the specific endpoint. Parameters are as in the endpoint wide response parameters described abovess.

