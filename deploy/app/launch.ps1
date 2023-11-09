param (
    [string]$portWeb = "8080",
    [string]$portBack = "8081",
    [switch]$f
 )

 # build front container
cd ..\..\APP
docker build . -t ogree-app
$assetsDir = "${PWD}\assets\custom"
$file = "${assetsDir}\.env"
(Get-Content $file) -replace '8081', $portBack | Set-Content $file

# run container
$basename = "ogree-superadmin"
$containername = $basename
$index = 1
While ($null -ne (docker ps --all --format "{{json .}}" --filter "name=$containername"))
{
    Write-Host "Container $containername already exists"
    if ($f.IsPresent) {
        Write-Host "Stopping it if running"
        docker stop $containername
    }
    $containername = "$basename-$index"
    $index++
}

Write-Host "Launch $containername container"
docker run --restart always --name $containername -p ${portWeb}:80 -v ${assetsDir}:/usr/share/nginx/html/assets/assets/custom -d ogree-app:latest
if ($LASTEXITCODE -ne 0) {
    Write-Host "UNABLE TO LAUNCH WEBAPP CONTAINER, CHECK ERROR ABOVE" -ForegroundColor red
}

# compile and run back
cd ..\BACK\docker-backend
docker run --rm -v ${PWD}:/workdir -w /workdir -e GOOS=windows golang go build -o ogree_app_backend.exe
.\ogree_app_backend.exe -port $portBack