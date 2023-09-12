#!/bin/bash
#
# Helper script catches environment variables passed from docker to init MongoDB.

mongosh localhost:27017 /home/createdb.js --eval '
var DB_NAME ="'$DB_NAME'",
ADMIN_DB="'$ADMIN_DB'",
CUSTOMER_API_PASSWORD=`'$CUSTOMER_API_PASSWORD'`,
SUPER_USER="'$SUPER_USER'",
SUPER_PASS="'$SUPER_PASS'",
ADMIN_USER="'$MONGO_INITDB_ROOT_USERNAME'",
ADMIN_PASS="'$MONGO_INITDB_ROOT_PASSWORD'",
GUARD_USER="'$GUARD_USER'",
GUARD_PASS="'$GUARD_PASS'"'
