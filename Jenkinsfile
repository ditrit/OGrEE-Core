pipeline {
    agent any

    stages {
        stage('SonarQube analysis') {
            environment {
              SCANNER_HOME = tool 'SonarQube-scanner'
            }
            steps {
            withSonarQubeEnv(credentialsId: 'OGREE-CLI', installationName: 'sonarqube-netbox') {
                 sh '''$SCANNER_HOME/bin/sonar-scanner \
                 -Dsonar.projectKey=OGREE-CLI \
                 -Dsonar.projectName=OGREE-CLI '''
               }
             }
        }

        stage('SQuality Gate') {
                steps {
                  timeout(time: 2, unit: 'MINUTES') {
                  waitForQualityGate abortPipeline: true
                  }
             }
        }

        stage('Build') {
            steps {
                echo 'Building..'
                sh '/OGrEE/buildService/updateCLI.py'


            }
        }

    }
}