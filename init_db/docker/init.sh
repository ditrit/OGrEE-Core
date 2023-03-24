#!/bin/bash
#
# Helper script catches environment variables passed from docker to init MongoDB.

mongosh localhost:27017 /home/dbft.js --eval '
var DB_NAME ="'$DB_NAME'",
CUSTOMER_API_PASS="'$CUSTOMER_API_PASS'",
CUSTOMER_RECORDS_DB="'$CUSTOMER_RECORDS_DB'",
ADMIN_DB="'$ADMIN_DB'",
SUPER_USER="'$SUPER_USER'",
SUPER_PASS="'$SUPER_PASS'",
ADMIN_USER="'$MONGO_INITDB_ROOT_USERNAME'",
ADMIN_PASS="'$MONGO_INITDB_ROOT_PASSWORD'",
GUARD_USER="'$GUARD_USER'",
GUARD_PASS="'$GUARD_PASS'"'
