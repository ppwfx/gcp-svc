# user-svc

## specification

### commands

The `user-svc` binary provides the following commands:

- `serve` serves the service
- `migrate` migrate the database

### security

- configuration
    - the service reads secrets from files

- communication
    - the loadbalancer terminates a https connection
    - a client can create a user account
    - a client authenticates a user account by providing the email, and password of the user
    - the service
        - salts and hashes the password
        - retrieves the user from the database
        - it validates the email and password combination
        - upon successful validation, the service
            - returns a JWT token that specifies
                - a `sub` claim, containing the email address of the user
                - a `exp` claim, containing a timestamp, that is 24 hours in the future
    - a client specifies a `Authorization: Bearer <token>` header that contains a JWT token
    - the service authorizes access to protected routes to JWT tokens, which
        - `sub` claim contains an email address, that ends with `@test.com`
        - `exp` claim contains a timestamp, that is in the future

- persistence
    - the service salts passwords, hashes the salted passwords, and stores the hashed passwords in the database

### testing

The test environment is setup with docker compose. The test environment consists of the following container:
    - `db`, hosts a postgres database
    - `user-svc`, hosts the service
    - `test`, migrates the database, and executes an integration test against `user-svc`

### database

The service uses a postgres database for persistence.

#### tables

The database provides the following tables:

- users
    - id (primary key, int, autoincrement)
    - email (unique, string)
    - fullname (string)
    - password (string)

#### migration

In the production context, `user-svc migrate` migrates the database

In the testing context, the test migrates the database

### communication

The service accepts, and returns `application/json` encoded objects via HTTP

#### routes

- api/v0/createUser
    - validation
        - email
            - is required
            - is email
            - doesn't appear in the `users` table yet
        - password
            - is required
            - has a minimum length of 8
        - fullname
            - is required
    - status codes
        - 400 on decoding failure
        - 422 on validation failure
        - 500 on internal server error

- api/v0/deleteUser
    - protected
    - validation
        - email
            - is required
            - is email
            - does appear in the `users`
    - status codes
        - 400 on decoding failure
        - 401 on unauthorized access
        - 422 on validation failure
        - 500 on internal server error

- api/v0/listUsers
    - protected
    - status codes
        - 400 on decoding failure
        - 401 on unauthorized access
        - 422 on validation failure
        - 500 on internal server error

- api/v0/authenticate
    - validation
        - email
            - is required
            - is email
        - password
            - is required
            - is email
    - status codes
        - 400 on decoding failure
        - 422 on validation failure
        - 500 on internal server error