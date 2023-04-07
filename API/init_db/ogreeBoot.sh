#!/usr/bin/env bash
#
#Initialise MongoDB server.

# NOTE: 
# Make sure that the API's user 
# matches the customer DB 'DB_NAME'


#######################################
# Parse flags or use default values 
# if not present.
# Flags:
#   path: Path where Mongo will 
#         store records
#   port: DB port to be exposed
#   log: Path for Mongo log
#   name: Name of the customer/dev database 
#   CUSTOMER_RECORDS_DB: Name of the customer records keeping database   
#   ADMIN_DB: Name of the Admin database to create
#   SUPER_USER: Name of the Super user to create
#   SUPER_PASS: Super user password 
#   ADMIN_USER: Name of the Admin user to create
#   ADMIN_PASS: Admin user password
#   GUARD_USER: Name of the Guard user to create
#   GUARD_PASS: Guard user password


# Default values:
#   path: ./mdb
#   port: 27017
#   log: /mongod.log
#   DB_NAME: ogreeDevelop
#   CUSTOMER_RECORDS_DB: ogree    
#   ADMIN_DB: admin
#   SUPER_USER: super
#   SUPER_PASS: superpassword
#   ADMIN_USER: admin
#   ADMIN_PASS: adminpassword
#   GUARD_USER: guard
#   GUARD_PASS: adminpassword

#######################################
while test $# -gt 0; do
           case "$1" in
                -path)
                    shift
                    path=$1
                    shift
                    ;;
                -port)
                    shift
                    port=$1
                    shift
                    ;;
                -log)
                    shift
                    log=$1
                    shift
                    ;;
                -host)
                    shift
                    host=$1
                    shift
                    ;;
                -CUSTOMER_RECORDS_DB)
                    shift
                    CUSTOMER_RECORDS_DB=$1
                    shift
                    ;;
                -DB_NAME)
                    shift
                    DB_NAME=$1
                    shift
                    ;;
                -SUPER_USER)
                    shift
                    SUPER_USER=$1
                    shift
                    ;;
                -SUPER_PASS)
                    shift
                    SUPER_PASS=$1
                    shift
                    ;;
                -ADMIN_USER)
                    shift
                    ADMIN_USER=$1
                    shift
                    ;;
                -ADMIN_PASS)
                    shift
                    ADMIN_PASS=$1
                    shift
                    ;;
                -GUARD_USER)
                    shift
                    GUARD_USER=$1
                    shift
                    ;;
                -GUARD_PASS)
                    shift
                    GUARD_PASS=$1
                    shift
                    ;;
                *)
                   echo "$1 is not a recognized flag!"
                   return 1;
                   ;;
          esac
  done  


if [ -z "$path"   ];
then
    path="./mdb"
fi

if [ -z "$port" ];
then
    port=27017
fi

if [  -z "$log" ];
then
    log="./mongod.log"
fi

if [  -z "$host" ];
then
    host="localhost"
fi

if [  -z "$DB_NAME" ];
then
    DB_NAME="TenantName"
fi

if [ -z "$CUSTOMER_RECORDS_DB"   ];
then
    CUSTOMER_RECORDS_DB="ogree"
fi

if [ -z "$ADMIN_DB"   ];
then
    ADMIN_DB="admin"
fi

if [ -z "$SUPER_USER"   ];
then
    SUPER_USER="super"
fi

if [ -z "$SUPER_PASS"   ];
then
    SUPER_PASS="superpassword"
fi

if [ -z "$ADMIN_USER"   ];
then
    ADMIN_USER="admin"
fi

if [ -z "$ADMIN_PASS"   ];
then
    ADMIN_PASS="adminpassword"
fi

if [ -z "$GUARD_USER"   ];
then
    GUARD_USER="guard"
fi

if [ -z "$GUARD_PASS"   ];
then
    GUARD_PASS="adminpassword"
fi

echo "Path : $path";
echo "Port : $port";
echo "Log :  $log";
echo "Host : $host";
echo "DB_NAME : $DB_NAME";

#killall mongod
fuser -k $port/tcp
rm -rf "$log"
rm -rf "$path"/*
mkdir "$path"
mongod --dbpath "$path" --port $port --logpath "$log" --fork

#Initialise the customer record DB
mongosh "$host:"$port bootup.js --eval '
var host = "'$host':'$port'", 
CUSTOMER_RECORDS_DB="'$CUSTOMER_RECORDS_DB'",
ADMIN_DB="'$ADMIN_DB'",
SUPER_USER="'$SUPER_USER'",
SUPER_PASS="'$SUPER_PASS'",
ADMIN_USER="'$ADMIN_USER'",
ADMIN_PASS="'$ADMIN_PASS'",
GUARD_USER="'$GUARD_USER'",
GUARD_PASS="'$GUARD_PASS'"'

#The command below will create the new customer DB 
mongosh "$host:"$port"/"$DB_NAME ./createdb.js --eval '
var host = "'$host':'$port'",
DB_NAME = "'$DB_NAME'",
CUSTOMER_RECORDS_DB = "'$CUSTOMER_RECORDS_DB'",
ADMIN_USER = "'$ADMIN_USER'",
ADMIN_PASS = "'$ADMIN_PASS'"'


# Create API User to access customer DB
echo 
echo "Please type a new a password for the customer: "
read PASS
mongosh "$host:"$port createUser.js --eval '
let DB_NAME = "ogree'$DB_NAME'",
ADMIN_USER = "'$ADMIN_USER'",
ADMIN_PASS = "'$ADMIN_PASS'",
PASS = "'$PASS'",
host = "'$host':'$port'";'

echo "PASSED BOOTUP"

sudo fuser -k $port/tcp
mongod --dbpath "$path" --port $port --logpath "$log" --fork --auth
echo "PASSED RESTART"
exit

echo "done"