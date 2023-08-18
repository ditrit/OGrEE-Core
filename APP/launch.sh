#!/bin/bash

portWeb=8080
portBack=8082
while getopts w:b: flag
do
    case "$flag" in
        w) portWeb=${OPTARG};;
        b) portBack=${OPTARG};;
    esac
done


assetsDir="$(pwd)/ogree_app/assets/custom"
file="$assetsDir/.env"
docker build -f APP -t ogree-app
sed -i "s/8081/$portBack/g" $file
docker run -p $portWeb:80 -v $assetsDir:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest

cd ./ogree_app_backend
docker run --rm -v $(pwd):/workdir -w /workdir golang go build -o ogree_app_backend
./ogree_app_backend -port $portBack
