param (
    [string]$portWeb = "8080",
    [string]$portBack = "8081"
 )

cd ..\..\APP
docker build . -t ogree-app
$assetsDir = "${PWD}\assets\custom"
$file = "${assetsDir}\.env"
(Get-Content $file) -replace '8081', $portBack | Set-Content $file
docker run --restart always --name ogree-superadmin -p ${portWeb}:80 -v ${assetsDir}:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest

cd ..\BACK\docker-backend
docker run --rm -v ${PWD}:/workdir -w /workdir -e GOOS=windows golang go build -o ogree_app_backend.exe
.\ogree_app_backend.exe -port $portBack