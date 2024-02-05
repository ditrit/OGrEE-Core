#!/bin/bash

portWeb=8080
portBack=8081
forceStop=false
isMac=false
while getopts w:b:fm flag
do
    case "$flag" in
        w) portWeb=${OPTARG};;
        b) portBack=${OPTARG};;
        f) forceStop=true;;
        m) isMac=true;;
    esac
done

# build front container
cd ../..
assetsDir="$(pwd)/APP/assets/custom"
file="$assetsDir/.env"
docker build . -f APP/Dockerfile -t ogree-app
sed -i "s/8081/$portBack/g" $file

# run container
basename="ogree-superadmin"
containername=$basename
index=1
while [[ $(docker ps --all --format "{{json .}}" | grep $containername) ]]; do
    echo "Container $containername already exists"
    if $forceStop; then
        echo "Stopping it if running"
        docker stop $containername
    fi
    containername="$basename-$index"
    ((index++))
done


echo "Launch $containername container"
docker run --restart always --name $containername -p $portWeb:80 -v $assetsDir:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest

# compile and run back
cd BACK/app
if $isMac; then
    echo "Compiling backend for macOS"
    docker run --rm -v $(pwd):/workdir -w /workdir -e GOOS=darwin golang go build -o ogree_app_backend
else
    docker run --rm -v $(pwd):/workdir -w /workdir -e GOOS=linux golang go build -o ogree_app_backend
fi
./ogree_app_backend -port $portBack
