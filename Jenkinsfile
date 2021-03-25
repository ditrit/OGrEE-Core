pipeline {
    agent any
    //agent {dockerfile true}

    //First way
    //properties([pipelineTriggers([githubPush()])])

    //2nd Way
    pipelineTriggers{
        triggers {
            githubPush()
        }
    }
    

    stages {
        stage('Build') {
            steps {
                echo 'Building..'
                bash 'cd /home/ziad/api/p3'
                bash 'git checkout master'
                bash 'git pull origin master'
                bash 'go build main'
                //bash ''
                //cd /var/lib/jenkins/workspace/Job1prototypev2
                //docker build -t testingalpine .

            }
        }
        stage('Test') {
            steps {
                echo 'Testing..'
                //docker rm $(docker ps -aq)
                //docker run testingalpine sh -c "cd prototypev2 && go test -v ./..."
            }

            stage('Unit Tests') {

            }

            stage('Functional Tests') {
                
            }
        }
        stage('Deploy') {
            steps {
                echo 'Deploying....'
            }
        }
    }
}

//The URL below is a good place to start
// https://raw.githubusercontent.com
// /github/platform-samples/master/hooks/jenkins/jira-workflow/Jenkinsfile