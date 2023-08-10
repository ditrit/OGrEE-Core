# Ogree BFF (Back for Frontend) for API Routing and Aggregation

This is the Ogree BFF (Back for Frontend) service that handles API redirection and aggregation, primarily designed to streamline frontend-backend communication. It acts as a gateway for routing requests to various backend APIs and provides a unified interface for frontend applications.

## Configuration

Before you begin, set the following environment variables in a .env file or as part of your runtime:

- `ARANGO_PASSWORD`: The password for authentication.
- `BFF_PORT`: The port where BFF listenning

You also need a api.json file, that contains an array of `name` and `url` of all API you when to use.

Notice that this minimal configuration you need is an API 'objects':

Exemple of api.json:

```json
[
    {"name":"server", "url": "http://localhost:8080"},
    {"name":"objects", "url": "http://localhost:3001"}
]
```
## Installation and Execution

1. Clone this repository to your local machine.
2. Install any necessary dependencies if required.
3. Compile and run the BFF using the following command:

```bash
go run main.go
```
The server should start and be ready to accept requests.

## Endpoints

### Bindings Data between objects's API and another

- GET /api/devices/:apiName/:objects/:objectAttributes/:devicesAttributes : Retrieve the list of all devices in the database connect to apiName where the specific attributes of the objects match the specific attributes of devices



## API Documentation
You can explore the BFF documentation using Swagger UI. After starting the server, navigate to /docs in your web browser to access the interactive documentation and explore the available endpoints, request and response schemas, and even test the API directly from the documentation.

For example, if the API is running locally, you can access Swagger UI at:

```bash
http://localhost:<BFF_PORT>/docs
``````