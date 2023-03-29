#!/usr/bin/env bash
#
#Initialise MongoDB server.


#######################################
# Parse flags or use default values 
# if not present.
# Flags:
#   path: Path where Mongo will 
#         store records
#   port: DB port to be exposed
#   log: Path for Mongo log
#   name: Name of the database 

# Default values:
#   path: ./mdb
#   port: 27017
#   log: /mongod.log
#   name: ogree
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

if [  -z "$name" ];
then
    name="ogreeDevelop"
fi

echo "Path : $path";
echo "Port : $port";
echo "Log :  $log";
echo "Host : $host";
echo "Name : $name";


#killall mongod
fuser -k $port/tcp
rm -rf "$log"
rm -rf "$path"/*
mkdir "$path"
mongod --dbpath "$path" --port $port --logpath "$log" --fork --auth

#Initialise the customer record DB
#"myTester:xyz123@"
echo "HI WE ALL HERE"
mongosh "$host:"$port bootup.js
echo "HEY WE PASSED THE BOOTUP"

#The command below will execute the mongo script
#--authenticationDatabase 'admin' -u 'myTester' -p 'xyz123'
mongosh --username "myTester" --password "xyz123" "$host:"$port"/admin" ./createdb.js --eval 'var dbName = "'$name'"' 
echo "done"