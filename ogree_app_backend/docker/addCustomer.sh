#!/bin/bash
#
# This script is used to create a new customer, and
# is executed only after the 'init.sh' script
#
# It is expected that the user will execute this 
# script inside a docker container
#
# Or replace this entirely by invoking mongosh 
# directly from outside the container such as:
# docker exec containerName sh -c '/home/addCustomer.sh -NEW_CUSTOMER yourName -NEW_CUSTOMER_API_PASS yourPass'


#######################################
# Parse flags or use default values 
# if not present.
# Flags:
#   NEW_CUSTOMER: Name of the new customer
#   NEW_CUSTOMER_API_PASS: API Password for new customer API

# Default values:
#   port: 27017
#   name: ogree
#######################################
while test $# -gt 0; do
           case "$1" in
                -NEW_CUSTOMER)
                    shift
                    NEW_CUSTOMER=$1
                    shift
                    ;;
                -NEW_CUSTOMER_API_PASS)
                    shift
                    NEW_CUSTOMER_API_PASS=$1
                    shift
                    ;;
                *)
                   echo "$1 is not a recognized flag!"
                   echo "You should provide values for" \
                        "'NEW_CUSTOMER' and 'NEW_CUSTOMER_API_PASS'"
                   exit 
                   return 1;
                   ;;
          esac
  done  

if [  -z "$NEW_CUSTOMER" ];
then
    echo "Please provide a name for new customer as a flag 'NEW_CUSTOMER'"
    exit
fi

if [ -z "$NEW_CUSTOMER_API_PASS" ]
then
    echo "Please provide an API Password " \
         "for the new customer as a flag 'NEW_CUSTOMER_API_PASS'"
    exit 
fi 

mongosh localhost:27017 addCustomer.js --eval '
var DB_NAME ="'$NEW_CUSTOMER'",
CUSTOMER_API_PASS="'$NEW_CUSTOMER_API_PASS'",
CUSTOMER_RECORDS_DB="'$CUSTOMER_RECORDS_DB'",
ADMIN_USER="'$ADMIN_USER'",
ADMIN_PASS="'$ADMIN_PASS'"'
