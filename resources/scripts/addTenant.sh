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


mongo "localhost:"$port createdb.js --eval 'var dbName = "'$name'"'