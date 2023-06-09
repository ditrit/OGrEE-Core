cd ..
docker build -f .\APP\Dockerfile . -t ogree-app
docker run -p 8080:80 -d ogree-app:latest
cd .\APP\ogree_app_backend
docker run --rm -v ${PWD}:/workdir -w /workdir -e GOOS=windows golang go build -o ogree_app_backend.exe
.\ogree_app_backend.exe