param (
    [string]$portWeb = "8080",
    [string]$portBack = "8082"
 )

cd ..
docker build -f .\APP\Dockerfile . -t ogree-app
$assetsDir = "${PWD}\APP\ogree_app\assets\custom"
$file = "${assetsDir}\.env"
(Get-Content $file) -replace '8082', $portBack | Set-Content $file
docker run -p ${portWeb}:80 -v ${assetsDir}:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest

cd .\APP\ogree_app_backend
docker run --rm -v ${PWD}:/workdir -w /workdir -e GOOS=windows golang go build -o ogree_app_backend.exe
.\ogree_app_backend.exe -port $portBack