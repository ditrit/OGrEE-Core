# Notes for The API
Designed with JWT, MongoDB and CI tested using Jenkins and docker with a
script to start up MongoDB


Introduction
------------
This is an API interfacing with a MongoDB cluster for data centre management.


Building
------------
  ### NOTE If Building On Windows
  You can use ```wsl``` and invoke ```make win``` to generate a windows build, however, this isn't strictly necessary and can be built manually.

  #### Build manually
  Execute the following in a terminal at the root directory of this project:
  ```
  $GITHASH=(git rev-parse HEAD)
  $GITHASHDATE=(git show -s --format=%ai)
  $GITBRANCH=(git branch --show-current)
  $DATE=Get-Date
  go build -ldflags="-X  cli/controllers.BuildHash=$GITHASH -X cli/controllers.BuildTree=$GITBRANCH -X cli/controllers.BuildTime=$DATE -X cli/controllers.GitCommitDate=$GITHASHDATE" .\main.go
  ```

  #### Building via wsl

  To install wsl please follow the [instructions here](https://learn.microsoft.com/en-us/windows/wsl/install)
  
  Launch wsl and navigate to the directory containing the API and continue to follow the instructions below. 

  If you want to generate a windows binary you must execute ```make win```

The ReadMe assumes that you already have the **latest version** of GO installed and your Environment PATHS properly setup  
You must also get [MongoDB](https://docs.mongodb.com/manual/installation/)  
For BSD systems, GO can be installed from the respective ports  
For Linux, consult your respective Distribution docs  

[Otherwise you can follow instructions from the GO site](https://golang.org/doc/install)  
   
  Clone the API repository  
  Execute ```make``` It should automatically retrieve the necessary libraries. If not then execute the command below to obtain the following packages
  ```
  go get github.com/dgrijalva/jwt-go github.com/fsnotify/fsnotify github.com/gorilla/mux go.mongodb.org/mongo-driver github.com/joho/godotenv gopkg.in/check.v1 github.com/crypto golang.org/x/sys golang.org/x/text gopkg.in/ini.v1  
  ```  

   Execute make


Running
-------------
You can modify the port of the API in the .env file. This is the port that the API will use to listen for requests.
 - Navigate in your terminal to the ```init_db``` directory  
 - Execute the bash script ```ogreeBoot.sh```
 - Enter your password when the prompt asks you
 - Be sure to enter your user password and the desired the DB access password
 - Update your .env file ```db_user=myCompanyName``` and ```db_pass=dbAccessPassword```
 - Execute the binary ```main```

This .env file is not provided, so you must create it yourself. Here is an example of the ```.env``` file:
```
PORT = 3001
db_host = 0.0.0.0
db_port = 27017
db_user = ""
db_pass = ""
db = "TenantName"
token_password = thisIsTheJwtSecretPassword
signing_password = thisIsTheRBACSecretPassword
email_account = "test@test.com"
email_password = ""
reset_url = "http://localhost:8082/#/reset?token="
``` 

MongoDB
--------------------------
A document (JSON) Based Database. The current DB can be started using the **ogreemdb.sh** script found in the root dir. The most useful interface for the DB is to access the shell. If you have started the DB already using the script you can directly execute this command to access the DB Shell
```
mongo --shell
```

Running a MongoDB container
--------------------------
docker run --name mdb -v /home/ziad/mongoDir:/docker-entrypoint-initdb.d/ -p 27017:27017 -d mongo:latest

Swagger Docs
--------------------------

Swagger API spec is generated from comments in the code, specifically in the controllers. For details of the format of these comments visit <https://goswagger.io/use/spec.html>.

The generation is performed with the `swagger` command, which can be installed by following the instructions at <https://goswagger.io/install.html>.

To generate a new version of the spec use the command:

```bash
go generate
```

To run locally an http server with the generated spec use:

```bash
swagger serve -p 3003 --no-open ./swagger.json
OR with docker:
docker run -p 80:8080 -e SWAGGER_JSON=/swagger.json -v ./swagger.json:/swagger.json swaggerapi/swagger-ui
```
