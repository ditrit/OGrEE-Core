pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'go build main.go'
                sh 'mv ./.env ./.env.bak'
                sh 'cp ./resources/test/.env .'

            }
        }

        stage('Docker Test') {
            //This stage is useless
            steps {
                echo 'Building Docker Image & Testing..'
                sh 'docker rmi $(docker images --filter "dangling=true" \
                -q --no-trunc) || true'

                sh 'DOCKER_BUILDKIT=1 docker build -t testingalpine:dockerfile .'
                //sh 'docker run testingalpine:dockerfile sh -c \
                //"cd p3 && go test -v ./..."'

                /*docker run --mount type=bind,source="$(pwd)"/resources/,
                target=/home -it postman/newman:alpine run 
                '/home/Basic Functionality.postman_collection.json'*/
            }
        }

        stage('Unit Testing') {
            steps {
                //sh 'go test -v ./models/... ./utils/...'
                sh 'go test -v  ./utils/...'
                echo 'Unit....'
            }
        }

        stage('Regression Testing') {
            steps {
                //sh 'go test -cover ./models/... ./utils/...'
                sh 'go test -cover ./utils/...'
                echo 'Regression....'
            }
        }

        stage('Functional Test') {
            steps {
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE'){
                echo 'Functional....'
                sh 'docker stop lapd || true'
                //sh 'cd ./resources/test && docker build -t apitester:dockerfile .'
                
                sh 'docker run --rm --network=roachnet -p 27018:27017 --name lapd -d -v /home/ziad/testMDB:/docker-entrypoint-initdb.d/ mongo'
                sh 'sleep 1'
                sh 'sudo ./main &'
                //script {
                //    
                //  env.RES = sh(script: 'sudo ./resources/test/scenario1.py || true', returnStdout: true).trim()
                //    
                //    
                //echo "RES = ${env.RES}"
                //}

                
                    sh 'sudo ./resources/test/scenario1.py'
                    sh 'sudo ./resources/test/scenario2.py'
                    sh 'sudo ./resources/test/scenario3.py'
                    sh 'sudo ./resources/test/scenario4.py'
                
                

                
                sh 'mv ./.env.bak ./.env'
                
                //sh 'docker run -d --rm --network=roachnet --name=rotten_apple_test testingalpine:dockerfile /bin/sh -c /home/main'
                //sh 'docker run -d --rm --network=roachnet --name=tester apitester:dockerfile /home/scenario1.py'
                //sh 'docker logs -f rotten_apple_test'
                //sh 'docker logs -f tester'
                //sh 'docker stop rotten_apple_test || true'
                //sh 'docker stop lapd || true'
                }
            }
        }

        //Generate binaries for other systems
        //and copy files 
        stage('Application Builds') {
            steps {
                //Linux Native
                sh 'go build -o OGrEE_API_Linux_x64 main.go'
                sh 'mv OGrEE_API_Linux_x64 /home/ziad/bin/api'

                //Windows x64
                sh 'GOOS=windows GOARCH=amd64 go build -o OGrEE_API_Win_x64 main.go'
                sh 'mv OGrEE_API_Win_x64 /home/ziad/bin/api'

                //OSX x64
                sh 'GOOS=darwin GOARCH=amd64 go build -o OGrEE_API_OSX_x64 main.go'
                sh 'mv OGrEE_API_OSX_x64 /home/ziad/bin/api'

                //Upload builds to Nextcloud
                sh '/home/ziad/buildServices/updateAPI.py'

                //OSX arm64
                //sh 'GOOS=darwin GOARCH=arm64 go build -o OGrEE_API_OSX_arm64 main.go'
                //sh 'mv OGrEE_API_OSX_arm64 /home/ziad/bin/api'

                //sh 'cp ./createdb.js /home/ziad/mongoDir'
            }
        }

        stage('Update development section') {
            steps {
                echo 'Deploying Development containers....'
                //Make backups of the DBs before stopping them
                sh '(docker exec cDB sh -c \'exec mongodump -d ogree --archive\' > /home/ziad/other/externalDBs/cDB/collection.archive) || true'
                sh '(docker exec hDB sh -c \'exec mongodump -d ogree --archive\' > /home/ziad/other/externalDBs/hDB/collection.archive) || true'
                sh '(docker exec zDB sh -c \'exec mongodump -d ogree --archive\' > /home/ziad/other/externalDBs/zDB/collection.archive) || true'
                sh '(docker exec tDB sh -c \'exec mongodump -d ogree --archive\' > /home/ziad/other/externalDBs/tDB/collection.archive) || true'
                sh '(docker exec vDB sh -c \'exec mongodump -d ogree --archive\' > /home/ziad/other/externalDBs/vDB/collection.archive) || true'

                //Restart services
                sh 'docker-compose -f /home/ziad/other/externalDBs/docker-compose.yml down || true'
                sh 'docker-compose -f /home/ziad/other/externalDBs/docker-compose.yml up -d'

                //Restore backups
                sh '(docker exec -i cDB sh -c \'exec mongorestore --archive\' < /home/ziad/other/externalDBs/cDB/collection.archive) || true'
                sh '(docker exec -i hDB sh -c \'exec mongorestore --archive\' < /home/ziad/other/externalDBs/hDB/collection.archive) || true'
                sh '(docker exec -i zDB sh -c \'exec mongorestore --archive\' < /home/ziad/other/externalDBs/zDB/collection.archive) || true'
                sh '(docker exec -i tDB sh -c \'exec mongorestore --archive\' < /home/ziad/other/externalDBs/tDB/collection.archive) || true'
                sh '(docker exec -i vDB sh -c \'exec mongorestore --archive\' < /home/ziad/other/externalDBs/vDB/collection.archive) || true'

               
            }
        }

        stage('Deploy') {
            steps {
                echo 'Deploying....'
                sh 'docker rmi $(docker images --filter "dangling=true" \
                -q --no-trunc) || true'

                sh 'DOCKER_BUILDKIT=1 docker build -t testingalpine:dockerfile .'
                sh 'docker stop lapd || true'
                sh 'fuser -k 27020/tcp || true'
                sh 'sudo fuser -k 3001/tcp || true'
                sh 'docker stop ogree_api || true'
                sh 'docker rm ogree_api || true'
                //sh 'rm ./env'
                //sh 'mv ./.env.bak ./.env'

                //Until we can figure out why the development section is
                //polluting the disk with volumes
                //juste delete them here in pipeline
                sh 'docker volume prune -f' 

                //Leftover containers from Dockerfile building
                sh 'docker container prune -f' 
                
                //sh 'docker run -d --rm --network=host --name=ogree_api testingalpine:dockerfile /home/main'
                sh 'docker-compose -f /home/ziad/api/v4/OGREE-APIv3/docker-compose.yml up -d --no-recreate'
                sh 'docker logs ogree_api'
               
            }
        }
    }
}