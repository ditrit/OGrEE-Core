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