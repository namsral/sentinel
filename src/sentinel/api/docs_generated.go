//Do not edit this file, it is generated.
package api

const (
docsRaml = `#%RAML 0.8

title: Sentinel API
baseUri: http://sentinel.sh/api/{version}
version: v1
documentation:
  - title: Signup
    content: |
      Signup an user using a valid email address and password to create a
      Sentinel account. After signup an email message is sent to the given
      email address to verify if the user controls the given email address.
  - title: Authenticate
    content: |
      Some methods of the API require the user to be authenticated.
      This is done by requesting an authentication token using the user's
      credentials, email and password. This authentication token is a JWT
      and has an expire date, usually an hour or so in which a new
      authentication token needs to be requested.
  - title: Authentication token
    content: |
      An authentication token is best described as a stateless session id with
      an expiry date.
  - title: One time login
    content: |
      A one time login allows a user to authenticate without a password. This
      can be used to change a forgotten password, to securely login on a
      public WiFi.
  - title: JWT
    content: |
      JSON Web tokens (JWT for short) are JSON objects signed by the Sentinel API.
      The tokens can contain information about resources managed by the Sentinel
      API and can be used to authenticate users, verify email addresses,
      password-less logins and more in the future. JWTs generated by the
      Sentinel API can be verified by clients using the public key available
      in the API.
mediaType: application/json; chartset=utf-8
traits:
  - secured:
      usage: Apply this to any method that needs to be secured
      description: Some requests require authentication.
      headers:
        Authorization:
          type: string
          example: Bearer eyJ...
      responses:
        401:
          headers:
            WWW-Authenticate:
              type: string
              example: |
                Bearer realm="https://Sentinel", error="invalid_credentials",
                error_description="missing or invalid authentication credentials"
          body:
            application/json; charset=utf-8:
              schema: error
              example: |
                {
                  "error": "invalid_credentials",
                  "error_description": "missing or invalid authentication credentials"
                }
        403:
          description: Unauthorized access.
  - limited:
      usage: 
      description: |
        Limit the range of returend results using HTTP Range headers as defined
        in [RFC6902](http://tools.ietf.org/html/rfc6902).
      headers:
        Range:
          description: |
            Set the limit and offset in a request by setting the range in the
            format 'first-last' or 'first-'. Index starts at 0.
          type: string
          example: 0-10
      responses:
        200:
          headers:
            Content-Range:
              description: |
                The response range as set by the server in the format
                'first-last/length' where length is the total number of
                results. The length can be set as a wildcard to indicate
                that the length is unknown or very large.
              type: string
              example: 0-10/20
schemas:
  - user: |
      { "$schema": "http://json-schema.org/schema",
        "type": "object",
        "properties": {
          "id": {
            "description": "UUID version 4 identifier as defined in RFC4122.",
            "type": "string",
            "format": "uuid"
          },
          "name": {
            "type": "string"
          },
          "lastLogin": {
            "type": "date"
          },
          "defaultAuthLevel": {
            "description": "Default authentication level, options are 0:unknown 1:notify 2:fast 3:secure",
            "type": "integer",
            "default": 0,
            "enum": [ 0, 1, 2, 3 ]
          },
          "deviceToken": {
            "description": "An APN device token or GCM registration token.",
            "type": "string",
            "maxLength": "256",
            "minLength": "64"
          }
        }
      }
  - authemail: |
      { "$schema": "http://json-schema.org/schema",
        "type": "object",
        "properties": {
          "id": {
            "description": "UUID version 4 identifier as defined in RFC4122.",
            "type": "string",
            "format": "uuid"
          },
          "email": {
            "type": "string",
            "pattern": "^[^@\s]+@[^@\s]+$"
          },
          "isVerified": {
            "description": "True when the email address is verified by the owner.",
            "type": "boolean",
            "default": false
          }
        }
      }
  - service: |
      { "$schema": "http://json-schema.org/schema",
        "type": "object",
        "properties": {
          "id": {
            "description": "UUID version 4 identifier as defined in RFC4122.",
            "type": "string",
            "format": "uuid"
          },
          "serviceUrl": {
            "type": "string",
            "format": "uri"
          },
          "serviceLogoUrl": {
            "type": "string",
            "format": "uri"
          },
          "authLevel": {
            "description": "Authentication level, options are 0:unknown 1:notify 2:fast 3:secure",
            "type": "integer",
            "default": 0,
            "enum": [ 0, 1, 2, 3 ]
          },
          "lastEntryDate": {
            "type": "date"
          }
        }
      }
  - error: |
      { "$schema": "http://json-schema.org/schema",
        "type": "object",
        "properties": {
          "error": {
            "type": "string"
          },
          "error_description": {
            "type": "string"
          }
        }
      }
/signup:
  post:
    description: |
      Signup for a Sentinel account with email and password. To verifiy the email
      address, an email message will be sent with a verification link.
    body:
      application/x-www-form-urlencoded; chartset=utf-8:
        formParameters:
          email:
            description: A valid email address controlled by the user.
            type: string
            pattern: ^[^@\s]+@[^@\s]+$
          password:
            description: The password with which the user registered the account.
            type: string
            minLength: 8
    headers:
      Prefer:
        description: Request the API to return the created resource
        type: string
        example: return=representation
    responses:
      201:
        description: |
          The response body is empty; to have the resource returned, include
          the the "Prefer: return=representation" header in the request.
        body:
          application/json; charset=utf-8:
            schema: user
      422:
        description: |
          Request had validation errors, email didn't match
          pattern, email already registered, etc.
        body:
          application/json; chartset=utf-8:
            schema: error
/token:
  post:
    description: Authenticate with email address and password to request an authentication token.
    headers:
      Authorization:
        type: string
        example: Basic dXNlcjpwYXNz
    responses:
      200:
        body:
          application/json; charset=utf-8:
            schema: |
              { "$schema": "http://json-schema.org/schema",
                "type": "object",
                "properties": {
                  "token_type": { "type": "string" },
                  "expires_in": { "type": "int" },
                  "id_token": { "type": "string" }
                }
              }
            example: |
              {
                "token_type": "Bearer",
                "expires_in": 3600,
                "id_token": "eyJ..."
              }
/onetimelogin:
  post:
    description: |
      Request a one time login. An email message with a one time loging link
      will be sent to the registered email address. This will allow the user
      to be authenticated with any credentials for a single, limited time.
    body:
      application/x-www-form-urlencoded; chartset=utf-8:
        formParameters:
          email:
            description: An email address with which the user registered an account.
            type: string
            pattern: ^[^@\s]+@[^@\s]+$
/user/self:
  is: [ secured ]
  get:
    description: Get user details for authenticated user.
    body:
      application/json; charset=utf-8:
        schema: user
  put:
    description: Update user details for authenticated user.
    body:
      application/x-www-form-urlencoded; chartset=utf-8:
        formParameters:
          name:
            description: Name of the user.
            type: string
          password:
            description: A new password.
            type: string
            minLength: 8
          defaultAuthLevel:
            description: |
              Default authentication level, options are 1:Notify, 2:Fast or 3:Secure.
            type: int
            enum: [ 1, 2, 3 ]
    responses:
      200:
        body:
          application/json; charset=utf-8:
            schema: user
      422:
        description: |
          Request had validation errors.
        body:
          application/json; chartset=utf-8:
            schema: error
/email:
  is: [ secured ]
  get:
    is: [ limited ]
    description: List all email addresses associated by the authenticated user.
    responses:
      200:
        body:
          application/json; charset=utf-8:
  post:
    description: |
      Register a new email address; a verification email will be sent to given
      email address.
    body:
      application/x-www-form-urlencoded; chartset=utf-8:
        formParameters:
          email:
            description: An email address with which the user registered an account.
            type: string
            pattern: ^[^@\s]+@[^@\s]+$
    headers:
        Prefer:
          description: Request the API to return the created resource
          type: string
          example: return=representation
    responses:
      201:
        description: |
          The response body is empty; to have the resource returned, include
          the the "Prefer: return=representation" header in the request.
        body:
          application/json; charset=utf-8:
            schema: authemail
      422:
        description: |
          Request had validation errors, email didn't match
          pattern, email already registered, etc.
        body:
          application/json; chartset=utf-8:
            schema: error
  /{id}:
    is: [ secured ]
    get:
      description: Get the email address.
      responses:
        200:
          body:
            application/json; chartset=utf-8:
              schema: authemail
    delete:
      description: Delete the email address.
      responses:
        204:
/verify:
  post:
    body:
      application/x-www-form-urlencoded; chartset=utf-8:
        formParameters:
          token:
            description: The base64 encoded JWT token which was sent in the verification email.
            type: string
            example: eyJ...
    responses:
      204:
      422:
        description: invalid token
        body:
          application/json; chartset=utf-8:
            schema: error
/service/{id}:
  get:
    description: Get the service associated with the id.
    responses:
      200:
        body:
          application/json; charset=utf-8:
            schema: service
/pubkey:
  get:
    description: Use this public key to validate the signature of JWT tokens created by the API.
    responses:
      200:
        body:
          text/plain; charset=utf-8:
`
)
