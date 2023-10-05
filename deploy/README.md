# How to deploy OGrEE

## Docker

In the `docker/` folder, you will find a docker compose file that can deploy a tenant with the following components: API, Mongo DB, WEB APP and Swagger UI DOC. The API and DB are the most essential part of OGrEE-Core, so they are always deployed. For the others, are optional.

To deploy only the api, in the docker folder, run:

```bash
docker compose -p mytenant up -d
```

To deploy all components, in the docker folder, run:

```bash
docker compose -p mytenant --profile web --profile doc up -d
```

Due to docker limitation, the tenant name passed in the -p argument (can also be configured in the .env file as `COMPOSE_PROJECT_NAME`) must contain only characters from [a-z0-9_-] and start with [a-z0-9].

The above commands can be easily executed using make:

```bash
make api
```

and

```bash
make all
```

respectively.

A .env file is present in the docker folder to configure the port the API, Web App and Swagger UI Doc (API_PORT, WEB_PORT, API_DOC_UI_PORT) will use, the OGrEE-Core location with CORE_DIR (local path or repo URL), the API external URL (API_DOC_UI_PORT) to be used by the Web App (if none, use localhost) and the password used by the API to authenticate to the Mongo DB (CUSTOMER_API_PASSWORD). The remaining DIR variables are the names of each subfolder in the Core repo.

### Dev version

There is also a development version with the following features:

* The MongoDB database is accessible on the `DB_PORT` (27017 by default) port. This will allow you to connect to the database using mongosh or mongodb-compass using the following url: `mongodb://ogreemytenantAdmin:pass123@localhost:27017/ogreemytenant?authSource=ogreemytenant&directConnection=true`.

This version can be launched by running:

```bash
make dev_api
```

## Manually

You can also manually compile and generate binaries of each component (see each component's readme for information on how). To create a database, you must first have Mongo installed. Then, run the `db/ogreeBoot.sh` script.

## Deploy APP+Docker Backend

To quickly deploy a frontend and docker backend in SuperAdmin mode, just execute the launch script appropriate to your OS from the `app` folder. This will use docker to compile both components and to run the frontend, the backend will be run locally.

```console
# Windows (use PowerShell)
.\launch.ps1
# Linux 
./launch.sh
```

This will launch the webapp on port 8080 and backend on port 8082. To set different ports:

```console
# Windows (use PowerShell)
.\launch.ps1 -portWeb XXXX -portBack YYYY
# Linux 
./launch.sh -w XXXX -b YYYY
```
