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
                -name)
                    shift
                    name=$1
                    shift
                    ;;
                *)
                   echo "$1 is not a recognized flag!"
                   return 1;
                   ;;
          esac
  done  


if [ -z "$port" ];
then
    port=27017
fi

if [  -z "$name" ];
then
    name="ogreeDevelop"
fi

echo "Port : $port";
echo "Name : $name";

#Create the secured Database
mongosh "localhost:"$port createdb.js --eval 'var dbName = "ogree'$name'"'


#Create an API user for the new customer
echo 
echo "Please type a new a password for the customer: "
read PASS
mongosh "localhost:"$port createUser.js --eval 'let dbName = "ogree'$name'", pass = "'$PASS'";'


#Success so print credentials one last time
echo "Great, be sure to save these credentials in your API env file" 
echo "since they will not be saved anywhere else! "
echo "db_user='$name'"
echo "db_pass='$PASS'"