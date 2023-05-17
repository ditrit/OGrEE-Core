# How to deploy OGrEE

## Docker
In the docker folder, you will find a docker compose file that can deploy a tenant with the following components: API, Mongo DB, WEB APP and Swagger UI DOC. The API and DB are the most essential part of OGrEE-Core, so they are always deployed. For the others, a profile should be passed. To build and run all componentes, in the docker folder, run:
```
docker compose -p mytenant --profile web --profile doc up -d
```
Due to docker limitation, the tenant name passed in the -p argument (can be also put in the .env file as COMPOSE_PROJECT_NAME) must contain only characters from [a-z0-9_-] and start with [a-z0-9].

A .env file should be present in the docker folder to set the port the API, Web App and Swagger UI Doc (API_PORT, WEB_PORT, API_DOC_UI_PORT) should use, the OGrEE-Core location with CORE_DIR (local path or repo URL), the API external URL (API_DOC_UI_PORT) to be used by the Web App (if none, use localhost) and the password used by the API to authenticate to the Mongo DB (CUSTOMER_API_PASSWORD). The remaining DIR variables are the names of each subfolder in the Core repo. Example of .env file:
```
# Local Core repo
CORE_DIR=../..
# Distant Core repo
# CORE_DIR=https://github.com/ditrit/OGrEE-APP.git#main
API_BUILD_DIR=API
APP_BUILD_DIR=APP
API_DOC_UI_PORT=8075
API_PORT=3001
WEB_PORT=8082
CUSTOMER_API_PASSWORD=pass123
API_EXTERNALURL=localhost
```

## Manually
You can also manually compile and generate binaries of each component (see each component's readme for information on how). To create a database, you must first have Mongo installed. Then, run the ogreeBoot.sh script located here. 