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

                docker run --mount type=bind,source="${pwd}"/resources/,target=/home \
                -it postman/newman:alpine run  \
                '/home/OGREED API.postman_collection.json || true'
            }
        }

        stage('Unit Testing') {
            steps {
                sh 'go test -v ./models/... ./utils/...'
                echo 'Unit....'
            }
        }

        stage('Regression Testing') {
            steps {
                sh 'go test -cover ./models/... ./utils/...'
                echo 'Regression....'
            }
        }

        stage('Functional Test') {
            steps {
                echo 'Functional....'
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