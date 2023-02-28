#!/usr/bin/env bash
#
#Initialise secured (MongoDB) DB for new tenant.


#######################################
# Parse flags or use default values 
# if not present.
# Flags:
#   port: DB port to be exposed
#   name: Name of the database 

# Default values:
#   port: 27017
#   name: ogree
#######################################
while test $# -gt 0; do
           case "$1" in
                -port)
                    shift
                    port=$1
                    shift
                    ;;
                -host)
                    shift
                    host=$1
                    shift
                    ;;
                -DB_NAME)
                    shift
                    DB_NAME=$1
                    shift
                    ;;
                -CUSTOMER_RECORDS_DB)
                    shift
                    CUSTOMER_RECORDS_DB=$1
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
                *)
                   echo "$1 is not a recognized flag!"
                   return 1;
                   ;;
          esac
  done  

if [  -z "$host" ];
then
    host="localhost"
fi

if [ -z "$port" ];
then
    port=27017
fi

if [  -z "$DB_NAME" ];
then
    echo "You need to provide the DB_NAME"
    exit
fi

if [  -z "$CUSTOMER_RECORDS_DB" ];
then
    echo "You need to provide the name of the CUSTOMER_RECORDS_DB"
    exit
fi

if [  -z "$ADMIN_USER" ];
then
    echo "You need to provide the ADMIN_USER credential"
    exit
fi

if [  -z "$ADMIN_PASS" ];
then
    echo "You need to provide the ADMIN_PASS credential"
    exit
fi

echo "Host : $host";
echo "Port : $port";
echo "Name : $DB_NAME";


#The command below will create the new customer DB 
mongosh "$host:"$port"/"$DB_NAME ./createdb.js --eval '
var host = "'$host':'$port'",
DB_NAME = "'$DB_NAME'",
CUSTOMER_RECORDS_DB = "'$CUSTOMER_RECORDS_DB'",
ADMIN_USER = "'$ADMIN_USER'",
ADMIN_PASS = "'$ADMIN_PASS'"'

#Create an API user for the new customer
echo 
echo "Please type a new a password for the customer: "
read PASS
mongosh "$host:"$port createUser.js --eval '
let DB_NAME = "ogree'$DB_NAME'",
ADMIN_USER = "'$ADMIN_USER'",
ADMIN_PASS = "'$ADMIN_PASS'",
PASS = "'$PASS'",
host = "'$host':'$port'";'


#Success so print credentials one last time
echo "Great, be sure to save these credentials in your API env file" 
echo "since they will not be saved anywhere else! "
echo "db_user='$DB_NAME'"
echo "db_pass='$PASS'"