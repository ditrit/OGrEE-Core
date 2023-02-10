//Check if host was passed as argument
//Otherwise use localhost
try {
  host;
} catch(e) {
  host = "localhost:27017"
}

//Authenticate first
var m = new Mongo(host)
var authDB = m.getDB("test")
authDB.auth('admin','adminpassword');


//Check if dbName was passed as argument
//Otherwise use "ogreeDevelop"
try {
  dbName;
} catch(e) {
  dbName = "ogreeDevelop"
}
var db = m.getDB(dbName)

try {
  pass;
} catch(e) {
  pass = "somethingElse" //code should not reach here
}
//passwordPrompt()

db.createUser({ user: dbName+"Admin", pwd: pass,
                roles: [{role: "readWrite", db: dbName}]
                })