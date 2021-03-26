pipeline {
    agent any
    //agent {dockerfile true}

    //First way
    //properties([pipelineTriggers([githubPush()])])

    //2nd Way
    /*pipelineTriggers{
        triggers {
            githubPush()
        }
    }*/
    

    stages {
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'go build main.go'
                //bash ''
                //cd /var/lib/jenkins/workspace/Job1prototypev2
                //docker build -t testingalpine .

            }
        }

        stage('Docker Test') {
            //This stage is useless
            steps {
                echo 'Building Docker Image & Testing..'
                //sh 'docker run testingalpine:dockerfile sh -c "cd p3 && go test -v ./..."'

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
                sh 'bash resources/startAPI.sh'
            }
        }
    }
}

//The URL below is a good place to start
// https://raw.githubusercontent.com
// /github/platform-samples/master/hooks/jenkins/jira-workflow/Jenkinsfile