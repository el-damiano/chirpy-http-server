# Basic HTTP Server - Chirpy

Minimalistic HTTP web server written in Go with basic data storage and a JSON
API.

*Built as a guided project on [Boot.dev](https://www.boot.dev/courses/learn-http-servers-golang)*

- [Install](#install)
- [Uninstall](#uninstall)
- [Usage](#usage)
  - [How to make a request](#how-to-make-a-request)
  - [How to use a response](#how-to-use-a-response)
  - [Available endpoints](#available-endpoints)
    - [Fileserver](#fileserver)
    - [Get ready status](#get-ready-status)
    - [Create user](#create-user)
    - [Login user](#login-user)
    - [Create post](#create-post)
    - [Retrieve all posts](#retrieve-all-posts)
    - [Retrieve post of specific UUID](#retrieve-post-of-specific-uuid)
    - [Page visits](#page-visits)
    - [Reset](#reset)

# Requirements

- postgresql (17.4+)
- go (1.24.0+)

# Install

1. Ensure `$GOPATH` is setup correctly, and add it to your `$PATH` in your
`.bashrc` / `.zshrc` like so:

```bash
export GOPATH=$(go env GOPATH)
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

2. Install it by running:

```bash
go install github.com/el-damiano/bootdev-http-server@latest
```

Or alternatively:

```bash
git clone https://github.com/el-damiano/bootdev-http-server.git &&
cd bootdev-http-server &&
go install .
```

# Uninstall

Either remove the file under `$GOBIN/bootdev-http-server`, or go to where you have
cloned the repo and run:

```bash
go clean -i -cache -modcache
```

# Usage

After installation is done, ensure that the `postgresql` service is running,
connect to it and create a database called `chirpy`.

Then create a `.env` file in your **current directory** and add your PostgreSQL
connection string. Format:

`
{protocol}://{username}:{password}@host:port/database
`

Sample of what I have:

`
DB_URL="postgres://{username}:{password}@localhost:5432/chirpy?sslmode=disable"
`

**NOTE**: `{username}` and `{password}` are not filled because it's something
you have to set up.

After all that just run it with `bootdev-http-server`. The URL will be
`localhost:8080`.

## How to make a request

All endpoints are accessible without any authentication nor authorization.

Requests contain the following elements:
- **Base URL**: in this case `localhost:8080`
- **Namespace**: indicates which group of requests you want to use
- **Method**: indicates which method from the group you want to use
- **Parameter**: singular, optional and only available to limited methods
- **JSON Payload**: optional and only available to limited methods

Request URL formats:

`{base_url}/{namespace}/{method}`

`{base_url}/{namespace}/{method}/{parameter}`

Sample request URL:

`localhost:8080/api/chirps`

Sample request URL with a parameter:

`localhost:8080/api/chirps/661492f5-5e2d-46dd-9b5e-44b6d65e434f`

## How to use a response

The only supported format is JSON.

Sample request and response:

`localhost:8080/api/chirps`

```bash
curl -X POST 'localhost:8080/api/users' -H 'Content-Type: application/json' -d '{"email": "john.pork@example.com"}'
```

```json
{
  "id": "38257b8b-7299-4baf-8824-8c40328f1023",
  "created_at": "2025-05-06T14:36:24.23833Z",
  "updated_at": "2025-05-06T14:36:24.23833Z",
  "email": "john.pork@example.com"
}
```

## Available endpoints

### Fileserver


`GET` `/app/`

Serves `index.html`.

### Get ready status

`GET` `/api/healthz`

Returns `200 OK` if server is ready.

### Create user

`POST` `/api/users`

Creates a user. Requires a JSON payload with an `email` and `password` values.
Returns JSON with `id`, `created_at`, `updated_at` and `email`.

Example usage:

```bash
curl -X POST 'localhost:8080/api/users' -H 'Content-Type: application/json' -d '{"email": "john.pork@example.com", "password": "superidoldexiaorong"}'
```

### Login user

`POST` `/api/login`

Logins user. Requires a JSON payload with `email` and `password`. Returns JSON
with `id`, `created_at`, `updated_at`, `email`, `token`, `refresh_token`.

`token` acts as your secret identifier when making requests to the API, so
don't lose or share it.

`refresh_token` acts as your token used to refresh your access `token`.

Example usage:

```bash
curl -X POST 'localhost:8080/api/login' -H 'Content-Type: application/json' -d '{"email": "john.pork@example.com", "password": "superidoldexiaorong"}'
```

### Update user information

`PUT /api/users`

Updates user information. Requires a JSON payload with the new `email` and
`password` and a header with access `token`. Returns JSON with `id`,
`created_at`, `updated_at`, `email`, `token`, `refresh_token`.

Example usage:

```bash
curl -i -X PUT 'localhost:8080/api/users' -H 'Content-Type: application/json' -H 'Authorization: Bearer <your access token here>' -d '{"email": "john.porkski@example.com", "password": "californiagirlswereunforgettable"}'
```

### Refresh token

`POST /api/refresh`

Generates an access `token`. Requires a header with `refresh_token`.

`refresh_token` acts as your token used to refresh your access `token`.

Example usage:

```bash
curl -i -X POST 'localhost:8080/api/refresh' -H 'Authorization: Bearer c3648f4f00b852ca3c428621120449f3eaae9e2255020b351025a7a345f0766c'
```

### Revoke token

`POST /api/revoke`

Revokes your `refresh_token`. Requires a header with `refresh_token`.

Example usage:

```bash
curl -i -X POST 'localhost:8080/api/revoke' -H 'Authorization: Bearer c3648f4f00b852ca3c428621120449f3eaae9e2255020b351025a7a345f0766c'
```

### Create post

`POST` `/api/chirps`

Creates a post. Requires a JSON payload with a `body` key and `Authorization`
header with a Bearer token. Returns JSON with `id`, `body`, `user_id`,
`created_at`, `updated_at` and `email`.

```bash
curl -X POST 'localhost:8080/api/chirps' -H 'Content-Type: application/json' -H 'Authorization: Bearer your-base64-token-here' -d '{"body": "hello I am John Pork"}'
```

### Retrieve all posts

`GET` `/api/chirps`

Retrieves all posts.

### Retrieve post of specific UUID

`GET` `/api/chirps/{chirpID}`

Retrieves a specific post. Requires a `{chirpID}` UUID parameter.

Example usage:

```bash
 curl -X GET 'localhost:8080/api/chirps/bdc22c7e-6cc5-424c-a07f-575b5a2d4b1b'
```

### Delete a post of specific UUID

`DELETE` `/api/chirps/{chirpID}`

Deletes a specific post. Requires a `{chirpID}` UUID parameter and an access
`token` Authorization Header.

Example usage:

```bash
curl -i -X DELETE 'localhost:8080/api/chirps/94b0bc50-f66b-49a0-ace0-2029a7e79622' -H 'Content-Type: application/json' -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiJiODVkZGU1YS1iNDg1LTQ5MDItODg4Yi05NDYyMjlkYjhiMDIiLCJleHAiOjQ2MzM1MzYxODgsImlhdCI6MTc0ODYzMDU2MX0.wMExxrAzje8pWLKYNXldJGjjFCsNMhxdQA73uYWNSZ0'
```

### Page visits

`GET` `/admin/metrics`

Serves HTML data containing page visit count of `/app/`

### Reset

**WARNING! IRREVERSIBLE!**
**Deletes all users and posts!**

*As a safety measure, requires `PLATFORM="dev"` set in `.env` file*

`POST` `/admin/reset`
