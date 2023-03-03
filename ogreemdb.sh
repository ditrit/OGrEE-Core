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

# Default values:
#   path: ./mdb
#   port: 27017
#   log: /mongod.log
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



echo "Path : $path";
echo "Port : $port";
echo "Log : $log";
echo "Host : $host";


#killall mongod
fuser -k $port/tcp
rm -rf "$log"
rm -rf "$path"/*
mkdir "$path"
mongod --dbpath "$path" --port $port --logpath "$log" --fork

#The command below will execute the mongo script
mongosh "$host:"$port"/ogree" ./init_db/createdb.js
echo "done"