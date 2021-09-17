#!/usr/bin/env bash

#https://www.mongodb.com/blog/post/mongodb-go-driver-tutorial
#https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo?utm_source=godoc
#killall mongod
fuser -k 27017/tcp
rm -rf ./mongod.log
rm -rf ./mdb/*
mkdir mdb
mongod --dbpath ./mdb --port 27017 --logpath ./mongod.log --fork

#The command below will execute the mongo script
mongo createdb.js
echo "done"