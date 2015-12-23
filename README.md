Sentinel
========

Sentinel is part of a service which enables its users to sign-in to third party services using an email address and their smartphone as a token of authentication. The Sentinel codebase was abandoned and serves only as a showcase for using two implementations:

	1. Go's Interfaces used as a constraint when designing standalone services, avoid code dependancy and mimic the decoupling in microservices.
	2. JSON Web Tokens used extensivly to replace web cookies and enhance security for both the service and the end-users.

The Sentinel API service can be built and made to run with a Postgres schema, but will be of little use without the 'missing' components like the smarthone apps.


Setup
-----

Setup the project to support package vendoring or use the `getgb.io` tool:

```sh
$ git clone https://github.com/namsral/sentinel
$ cd sentinel
$ export PROJECT=$(pwd) && export GOPATH=$PROJECT:$PROJECT/vendor
```

Install dependancies:

```sh
$ GOPATH=$PROJECT/vendor go get github.com/lann/squirrel
```

_You can also use the `gb` tool for this._


Build and install the `sentinel` binary:

```sh
$ go install sentinel/cmd/sentinel
```

_The binary will be installed to `$PROJECT/bin/`._


Setup the PostgreSQL database and export the PG environment:

```sh
$ export PGHOST=127.0.0.1
$ export PGDATABASE=sentinel
$ export PGUSER=sentinel
$ export PGPASSWORD=secret
```

Create the database tables etc:

```sh
$ sentinel createdb
```

Run the HTTP service:

```sh
$ sentinel serve
```


API Call Examples
-----------------

Singup with email and password:

```sh
$ curl -i -X POST -d "email=jake@example.com&password=ninja" http://localhost:6000/api/v1/user/signup
    HTTP/1.1 200 OK
    Content-Type: application/json; charset=utf-8
    Date: Mon, 22 Jun 2015 08:47:47 GMT
    Content-Length: 253

    {
      "access_token": "5b8af00b-18bb-11e5-9225-b8e8560f873a",
      "refresh_token": "5b8af018-18bb-11e5-9225-b8e8560f873a",
      "user_id": "5b88f178-18bb-11e5-9225-b8e8560f873a",
      "created_at": "2015-06-22T08:47:47.845122805Z",
      "expires_in": 3600000000000
    }
```

Get user details using token authentication:

```sh
$ curl -i -H "Authorization: Bearer 5b8af00b-18bb-11e5-9225-b8e8560f873a" http://localhost:6000/api/v1/user
    HTTP/1.1 200 OK
    Content-Type: application/json; charset=utf-8
    Date: Mon, 22 Jun 2015 08:49:14 GMT
    Content-Length: 227

    {
      "id": "5b88f178-18bb-11e5-9225-b8e8560f873a",
      "name": "",
      "lastLogin": "0001-01-01T00:00:00Z",
      "defaultAuthLevel": 0,
      "authEmailList": [
        {
          "email": "jake@example.com",
          "isVerified": false
        }
      ]
    }
```


API Documentation
-----------------

The Sentinel API service publishes its own documentation in RAML format on 'https://sentinel.sh/api/v1/docs'.

	#%RAML 0.8

	title: Sentinel.sh Login flow for 3rd Parties
	baseUri: http://sentinel.sh/api/{version}
	version: v1
	/qauth/login:
	  is: [ secured ]
	  description: |
		Create a login request for the user identified by their email
		address and the authenticated service provider. (step 2)
	  post:
		body:
		  application/json; charset=utf-8:
			description: |
			  The secret1 property is passed allong to the mobile device for it to
			  validate the login request.
			example:
			  {
				"email": "bob@example.com",
				"secret1": "ffa6706ff2127a749973072756f83c532e43ed02"
			  }
		responses:
		  200:
			description: |
			  Return a session ID with which the login request can be identified by
			  the 3rd party service in step 7.  
			body:
			  application/json; charset=utf-8:
				example: |
				  {
					"sessionID": "a5828e8b-b203-49ba-8aa0-60b9dfb20220"
				  }
	//gateway.sandbox.push.apple.com:
	  description: |
		Send push notification to mobile device associated with email address.
		The following example is extra data appended to the push notification
		payload. (step 3)
	  example:
		  {
			"email": "bob@example.com",
			"secret1": "ffa6706ff2127a749973072756f83c532e43ed02",
			"sessionID": "a5828e8b-b203-49ba-8aa0-60b9dfb20220",
			"service": {
			  "id": "a28c719d-5594-46a9-bfe7-97fa4d35e3fe",
			  "serviceUrl": "https://example.com",
			  "serviceLogoUrl": "https://cdn.example.com/i/logo.png"
			}
		  }
	//example.com/verify:
	  description: |
		The mobile device trades secrets with the 3rd party. (step 5)
	  post:
		body:
		  application/json; charset=utf-8:
			example:
			  {
				"secret1": "ffa6706ff2127a749973072756f83c532e43ed02",
				"sessionID": "a5828e8b-b203-49ba-8aa0-60b9dfb20220"
			  }
		responses:
		  200:
			body:
			  application/json; charset=utf-8:
				example:
				  {
					"secret2": "29df362b5cfa5c96d22f8d20f29d9a367dd0d359"
				  }
	/qauth/status:
	  is: [ secured ]
	  description: |
		Mobile device responds wether it accepts or declines the login request. (step 6)
	  post:
		body:
		  application/json; charset=utf-8:
			description: |
			  The token property is the same token from step 3 and includes email,
			  serviceURL, secret1, etc.
			  The enc1 property is gained from contact with the 3rd party example.com
			  service.
			  The secret2 property is acquired when trading secrets in step 5.
			example:
			  {
				"status": "accept",
				"token": "eyJ...",
				"secret2": "29df362b5cfa5c96d22f8d20f29d9a367dd0d359"
			  }
	//example.com/status:
	  description: Sentinel.sh calls the status endpoint with the appropriate data. (step 7)
	  post:
		body:
		  application/json; charset=utf-8:
			description: |
			  The sessionID property is the same ID from step 2 and can be used to
			  identiy the login request.
			example:
			  {
				"status": "accepted",
				"secret2": "29df362b5cfa5c96d22f8d20f29d9a367dd0d359",
				"sessionID": "a5828e8b-b203-49ba-8aa0-60b9dfb20220"
			  }
			example:
			  {
				"error": "login_timeout",
				"error_description": "login was not confirmed within default timeout"
			  }
		responses:
		  2xx:
			description: On a successful response the login request will be
			  completed, on error the callbak will be tried once again.

