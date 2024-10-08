# RESTful API Sample Application

This application provides a very basic Go RESTful web server intended to run on http://localhost:8080. The API writes data to in-memory caches which are wiped clean upon restart. The RESTful components of this API are generated from a OpenAPI 3.0 [document](docs/openapi.json), ensuring code stays in alignment with the documentation.

### Requirements

See [requirements.pdf](docs/requirements.pdf)

## Running Application
This project requires Go (see [go.mod](go.mod) for version) to build, run and test the application. You can download Go from [here](https://go.dev/dl/).

After confirming Go is installed, please run `make help` and review the make targets.  If you would like to run the application on a non-default address, you may do so by providing the command line arg `addr` like this: `go run cmd/rest/main.go --addr=localhost:1234`.

## Testing

There are two sets of tests for the application.  Due to time constraints, I've only provided test samples as the remaining are mostly boilerplate versions of what I've written.  Running `make test` will execute the unit tests.

Integration tests are using IntellJ's HTTP client.  Unless you already use Jetbrains products with this built-in, you'll need to install the [IJHTTP client](https://www.jetbrains.com/ijhttp/download) (requires Java 17).

> NOTE: Only test failures are reported by the CLI.

```
jason@Durham http % ~/Downloads/ijhttp/ijhttp -e aws -v ./http-client.env.json ./users.http ./posts.http
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Running IntelliJ HTTP Client with                      │
├──────────────────────┬──────────────────────────────────────────────────────┤
│        Files         │ users.http,                                          │
│                      │ posts.http                                           │
├──────────────────────┼──────────────────────────────────────────────────────┤
│  Public Environment  │ scheme = http://,                                    │
│                      │ addr = 52.23.252.97:8080                             │
├──────────────────────┼──────────────────────────────────────────────────────┤
│ Private Environment  │                                                      │
└──────────────────────┴──────────────────────────────────────────────────────┘
Request 'Fetch users' GET http://52.23.252.97:8080/users
Request 'Create a user' POST http://52.23.252.97:8080/users
Request 'Get a valid user' GET http://52.23.252.97:8080/users/4
Request 'Get a user that doesn't exist' GET http://52.23.252.97:8080/users/0
Request 'Update a user that exists' PUT http://52.23.252.97:8080/users/4
Request 'Update a user that doesn't exist' PUT http://52.23.252.97:8080/users/0
Request 'Delete user that exists' DELETE http://52.23.252.97:8080/users/4
Request 'Delete user that does not exist' DELETE http://52.23.252.97:8080/users/0
Request 'Create a user for tests' POST http://52.23.252.97:8080/users
Request 'Fetch posts' GET http://52.23.252.97:8080/posts
Request 'Create a post' POST http://52.23.252.97:8080/posts
Request 'Get a valid post' GET http://52.23.252.97:8080/posts/2
Request 'Get a post that doesn't exist' GET http://52.23.252.97:8080/posts/0
Request 'Update a post that exists' PUT http://52.23.252.97:8080/posts/2
Request 'Update a post to set a user that doesn't exist' PUT http://52.23.252.97:8080/posts/2
Request 'Delete a post that exists' DELETE http://52.23.252.97:8080/posts/2
Request 'Delete a post that does not exist' DELETE http://52.23.252.97:8080/posts/0
Request 'Delete user created for tests' DELETE http://52.23.252.97:8080/users/5


18 requests completed, 0 have failed tests
RUN SUCCESSFUL
```

## Makefile

See `make help` to learn how to run the application with default settings, execute tests, check for CVEs and execute 100+ linters.

## Deployment

This is deployed on a [small EC2 instance](http://52.23.252.97:8080/users) for demo purposes.  There is a `make` target to build and copy the binary to the demo server, but it requires SSH configuration that is not available in this repo.
