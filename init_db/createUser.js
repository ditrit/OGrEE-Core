DB_NAME;
ADMIN_USER;
ADMIN_PASS;
PASS;

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
authDB.auth(ADMIN_USER, ADMIN_PASS);



var db = m.getDB(DB_NAME)
db.createUser({ user: DB_NAME+"Admin", pwd: PASS,
                roles: [{role: "readWrite", db: DB_NAME}]
                })