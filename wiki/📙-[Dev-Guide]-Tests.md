
It is important to test our code to reduce the bugs and assure the code works properly. The OGrEE-Core project presents tests for its different components:

- [API](#api)
- [CLI](#cli)
- [APP](#app)

API
--------------------------
To test the API you will need a test database. This database can be created using the script defined in `deploy/docker`.

```
# At the root directory of OGrEE-Core project
cd deploy/docker
make test_api
```

Once the test database is running, you can execute the tests at the API directory


```
# At the API directory, run all the tests
go test -p 1 ./...

# To avoid using cache in the tests you can add the flag count
go test -p 1 -count=1 ./...
```

You can also generate a coverage report and search for the uncoveraged lines that need a test

```
# Run all the tests and generate the coverage report 
go test -p 1 -count=1 -coverpkg=./... -coverprofile=coverage/coverage.out ./...

# Output the coverage report
go tool cover -func coverage/coverage.out

# Create an html report that shows the coverage on each file
go tool cover -html=coverage/coverage.out
```

The API has a test directory that defines different functions that can be used during the tests.

- The `e2e` directory defines functions that are useful for end to end tests. this means, the tests that checks the different API endpoints. Here we have defined the function that simulates the API request (MakeRequest*). There are also some common functions that make a request, check some common validations (ValidateRequest*) and return the response. These were created to avoid code duplication.

- The `integration` directory defines the database connection with the test database and defines multiple functions that allow us to create different types of entities (such as Sites, rooms, etc) that will be used during the tests. It is possible to create temporary entities that will only exist during the test and will be deleted at the end of the test (for example CreateTestPhysicalEntity)

- The `unit` directory defines some useful functions

- The `utils` directory defines some useful tests functions related to the obtention of an endpoint or the obtention of the body of an entity

CLI
--------------------------
You can test the CLI by executing

```
# At the CLI directory, run all the tests
go test -p 1 ./...

# To avoid using cache in the tests you can add the flag count
go test -p 1 -count=1 ./...
```

To generate the coverage report we need to ignore the readline directory as it is an external package

```
# Run all the tests and generate the coverage report 
go test -p 1 -coverprofile=coverage/coverage.out `go list ./... | grep -v ./readline`

# Output the coverage report
go tool cover -func coverage/coverage.out

# Create an html report that shows the coverage on each file
go tool cover -html=coverage/coverage.out
```


Depending on what we are testing, the CLI may need to communicate with the API or another external service. For those cases, the `mocks` directory defines different mocks to simulate the external service behavior. 

- API: it is used to simulate the API during the tests. In the test, you can define the mock and the endpoints that it accepts, which is going to be its response and verify if it was called during the execution.

- Ogree3D: it is used to simulate the communication with the [OGrEE-3D](https://github.com/ditrit/OGrEE-3D) component. The CLI communicates with OGrEE-3D to send information when a change is made, allowing OGrEE-3D to updates its view if necessary.


APP
--------------------------
You can test the Flutter application by executing

```
# At the APP directory, run all the tests
flutter test
```

By adding the flag `coverage` flutter will generate a coverage report `coverage/lcov.info`. We can install `lcov` (which will add gethtml) and use it to generate an html version

```
# Run all the tests and generate the coverage report 
flutter test --coverage

# Create an html report that shows the coverage on each file
genhtml coverage/lcov.info -o coverage/html

# Open the file coverage/html/index.html with your browser
```
