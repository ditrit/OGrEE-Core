# ArangoDB Database API in GoLang

This GoLang API enables management of "devices" and "connections" elements within an ArangoDB database. It provides endpoints to perform CRUD (Create, Read, Update, Delete) operations on these elements.


## Configuration

Before you begin, ensure you have an ArangoDB database up and running. Set the following environment variables in a .env file or as part of your runtime environment:

- `ARANGO_URL`: The URL of your ArangoDB instance.
- `ARANGO_DATABASE`: The name of the ArangoDB database you'll be using.
- `ARANGO_USER`: The username for authentication.
- `ARANGO_PASSWORD`: The password for authentication.
- `ÀPI_PORT`: The port where API listenning


For example, you can set these environment variables in a `.env` file in the root directory of your project:

```env
ARANGO_URL=http://localhost:8529
ARANGO_DATABASE=mydatabase
ARANGO_USER=user
ARANGO_PASSWORD=password
API_PORT="8081"
```

## Installation and Execution

1. Clone this repository to your local machine.
2. Install any necessary dependencies if required.
3. Compile and run the API using the following command:

```bash
go run main.go
```
The server should start and be ready to accept requests.

## Endpoints
### Devices

- GET /devices : Retrieve the list of all devices.
- GET /devices/:id : Retrieve information about a specific device based on its ID.
- GET /devices/ConnectedTo/:id : Retrieve the list of all devices connected to a specific device
- POST /devices :Add a new device to the database. Data must be provided in the request body in JSON format.
- DELETE /devices/:id :Delete a specific device based on its ID.


Example JSON data for a new device:

```json
{
  "_name": "storage_bay",
  "category": "port",
  "created": "2016-04-22",
  "expired": "3000-01-01",
  "group_name": "GS00OPSAN06",
  "hba_device_name": "nsa.*",
  "sp_name": "sp_b",
  "sp_port_id": "0",
  "storage_group_name": "storage"
}
```

### Connections


- GET /connections : Retrieve the list of all connections.
- POST /connections : Add a new connection to the database. Data must be provided in the request body in JSON format.
- DELETE /connections/:id : Delete a specific connection based on its ID.


Example JSON data for a new connection:

```json
  {
  "_from": "devices/*",
  "_to": "devices/*",
  "created": "2016-04-22",
  "expired": "3000-01-01",
  "type": "parent of (between partens)"
}
```

### Responses
The API returns data in JSON format for all operations. Responses typically include a array containing the requested data or an error message in case of an issue.

Example successful response:

```json

{
[
{
  "_name": "storage_bay",
  "category": "port",
  "created": "2016-04-22",
  "expired": "3000-01-01",
  "group_name": "GS00OPSAN06",
  "hba_device_name": "nsa.*",
  "sp_name": "sp_b",
  "sp_port_id": "0",
  "storage_group_name": "storage"
}
]
}
```
Example error response:

```json
{
  "message": "Error message"
}
```

## API Documentation
You can explore the API documentation using Swagger UI. After starting the server, navigate to /docs in your web browser to access the interactive documentation and explore the available endpoints, request and response schemas, and even test the API directly from the documentation.

For example, if the API is running locally, you can access Swagger UI at:

```bash
http://localhost:8080/docs
``````