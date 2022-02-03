pipeline {
    agent any
    /*environment {
        DATE = '''${sh (
            script: 'date +%Y.%m.%d//%T',
            returnStdout: true).trim()}'''

        GITHASH = '''${sh(
            script: 'git rev-parse HEAD',
            returnStdout: true).trim()}'''

        GITBRANCH = '''${sh(
            script: 'git branch --show-current',
            returnStdout: true).trim()}'''

        GITHASHDATE = '''${sh(
            script: 'git show -s --format=%ci HEAD',
            returnStdout: true).trim()}'''

    }*/
    stages {
        stage('Build') {
            steps {
                echo 'Building..'
                //sh 'rm makefile'
                //sh 'cp  /home/ziad/buildCLI/makefile .'
                //sh 'make'
                sh '/home/ziad/buildCLI/updateCLI.py'

                //Code block for Jenkins building on Chibois


            }
        }

        stage('Unit Testing') {
            steps {
                //sh 'go test -v ./models/... ./utils/...'
                echo 'Unit....'
            }
        }


        //Generate binaries for other systems
        //and copy files 
        stage('Application Builds') {
            steps {
                sh 'docker build -t cli:latest .'
                echo 'done'

                //Linux Native
                /*sh 'go build -o OGrEE_CLI_Linux_x64 -ldflags="-X  cli/controllers.BuildHash=${GITHASH} -X cli/controllers.BuildTree=${GITBRANCH} -X cli/controllers.BuildTime=${DATE} -X cli/controllers.GitCommitDate=${GITHASHDATE}" main.go ast.go lexer.nn.go y.go repl.go'
                sh 'mv OGrEE_CLI_Linux_x64 /home/ziad/bin/cli'

                //Windows x64
                sh 'GOOS=windows GOARCH=amd64 go build -o OGrEE_CLI_Win_x64 -ldflags="-X  cli/controllers.BuildHash=${GITHASH} -X cli/controllers.BuildTree=${GITBRANCH} -X cli/controllers.BuildTime=${DATE} -X cli/controllers.GitCommitDate=${GITHASHDATE}" main.go ast.go lexer.nn.go y.go repl.go'
                sh 'mv OGrEE_CLI_Win_x64 /home/ziad/bin/cli'

                //OSX x64
                sh 'GOOS=darwin GOARCH=amd64 go build -o OGrEE_CLI_OSX_x64 -ldflags="-X  cli/controllers.BuildHash=${GITHASH} -X cli/controllers.BuildTree=${GITBRANCH} -X cli/controllers.BuildTime=${DATE} -X cli/controllers.GitCommitDate=${GITHASHDATE}" main.go ast.go lexer.nn.go y.go repl.go'
                sh 'mv OGrEE_CLI_OSX_x64 /home/ziad/bin/cli'*/

                //OSX arm64
                //sh 'GOOS=darwin GOARCH=arm64 go build -o OGrEE_CLI_OSX_arm64 main.go'
                //sh 'mv OGrEE_CLI_OSX_arm64 /home/ziad/bin/cli'

            }
        }

    }
}