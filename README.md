# RESTful API Sample Application

This application provides a very basic Go RESTful web server intended to run on http://localhost:8080. The API writes data to in-memory caches which are wiped clean upon restart. The RESTful components of this API are generated from a OpenAPI 3.0 [document](docs/openapi.json), ensuring code stays in alignment with the documentation.

### Requirements

See [requirements.pdf](docs/requirements.pdf)

## Running Application
This project requires Go (see [go.mod](go.mod) for version) to build, run and test the application. You can download Go from [here](https://go.dev/dl/).

After confirming Go is installed, please run `make help` and review the make targets.  If you would like to run the application on a non-default address, you may do so by providing the command line arg `addr` like this: `go run cmd/rest/main.go --addr=localhost:1234`.

## Testing

There are two sets of tests for the application.  Due to time constraints, I've only provided test samples as the remaining are mostly boilerplate versions of what I've written.  Running `make test` will execute the unit tests.  

Integration tests are using IntellJ's HTTP client.  This was a new endeavor for me, but I find them very similar to Postman.  It looks like you can run these tests using this [CLI](https://blog.jetbrains.com/idea/2022/12/http-client-cli-run-requests-and-tests-on-ci/) but I didn't have time to automate this for you.

## Makefile

See `make help` to learn how to run the application with default settings, execute tests, check for CVEs and execute 100+ linters.
