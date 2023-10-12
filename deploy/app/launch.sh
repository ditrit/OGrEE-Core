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

cd ../../APP
assetsDir="$(pwd)/assets/custom"
file="$assetsDir/.env"
docker build . -t ogree-app
sed -i "s/8081/$portBack/g" $file
docker run --restart always --name ogree-superadmin -p $portWeb:80 -v $assetsDir:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest

cd ../BACK/docker-backend
docker run --rm -v $(pwd):/workdir -w /workdir golang go build -o ogree_app_backend
./ogree_app_backend -port $portBack
