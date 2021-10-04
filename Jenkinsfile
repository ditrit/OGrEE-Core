pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'go build main.go'

            }
        }

        stage('Docker Test') {
            //This stage is useless
            steps {
                echo 'Building Docker Image & Testing..'
                sh 'docker rmi $(docker images --filter "dangling=true" \
                -q --no-trunc) || true'

                sh 'docker build -t testingalpine:dockerfile .'
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
                echo 'Functional....'
                sh 'docker stop lapd || true'
                sh 'docker run --name lapd -d -v /home/ziad/testMDB:/docker-entrypoint-initdb.d/ -p 27018:27017 mongo'
                sh 'docker run -d --rm --name=rotten_apple_test testingalpine:dockerfile /home/scenario1.py'
                sh 'docker logs -f rotten_apple_test'
                sh 'docker stop rotten_apple_test || true'
                sh 'docker stop lapd || true'
            }
        }

        stage('Deploy') {
            steps {
                echo 'Deploying....'

                sh 'docker stop rotten_apple || true'
                sh 'docker run -d --rm --network=host --name=rotten_apple testingalpine:dockerfile /home/main'
               
            }
        }
    }
}