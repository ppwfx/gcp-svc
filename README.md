# user-svc

## specification

### security

- configuration
    - the service reads secrets from files

- communication
    - the load balancer terminates a https connection
    - a client can create a user account
    - a client authenticates a user account by providing the users email, and password
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