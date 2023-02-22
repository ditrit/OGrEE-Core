#!/bin/bash
#
# Helper script catches environment variables passed from docker to init MongoDB.

mongosh localhost:27017 dbft.js --eval 'var DB_NAME ="'$DB_NAME'",
CUSTOMER_API_PASS="'$CUSTOMER_API_PASS'",
CUSTOMER_RECORDS_DB="'$CUSTOMER_RECORDS_DB'",ADMIN_DB="'$ADMIN_DB'",
SUPER_USER="'$SUPER_USER'",SUPER_PASS="'$SUPER_PASS'",ADMIN_USER="'$ADMIN_USER'",
ADMIN_PASS="'$ADMIN_PASS'",GUARD_USER="'$GUARD_USER'",
GUARD_PASS="'$GUARD_PASS'"'
