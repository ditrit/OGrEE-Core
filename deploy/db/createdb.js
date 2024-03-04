/////
// NOTE
//
// An 'admin' DB will be created with an admin, super and backup user
// MongoDB docker image will execute scripts in alphabetical order
//
// Finally a secured customer DB will be created with an API user
// credential
//
// It is recommended that you use a secure method for supplying 
// passwords
//////

// How to Authenticate
//
// As Admin:
// mongosh "mongodb://ADMIN_USER:ADMIN_PASS@localhost/test?authSource=test"
// 
// As Super:
// mongosh "mongodb://SUPER_USER:SUPER_PASS@localhost/test?authSource=test" 
//
// As API:
// mongosh "mongodb://"ogree"+DB_NAME+"Admin":CUSTOMER_API_PASSWORD@localhost/"ogree"+DB_NAME?authSource="ogree"+DB_NAME"

// CONSTANT DECLARATIONS
DB_NAME;
CUSTOMER_API_PASSWORD;

ADMIN_DB;
SUPER_USER;
SUPER_PASS;

ADMIN_USER;
ADMIN_PASS;

GUARD_USER;
GUARD_PASS;

var m = new Mongo()
var authDB = m.getDB(ADMIN_DB)

// Get all existing users
var users = authDB.getUsers()["users"];
var found = false;

// Check if a specific user exists 
// we loop here for future proofing purposes
// this is meant for docker-compose 
for (var i = 0; i < users.length; i++) {
    if (users[i].hasOwnProperty('user') && users[i]['user'] === ADMIN_USER) {
        console.log("User already exists, skip user creation")
        found = true;
    }
}

//Create users if not found
if (!found) {
    authDB.createUser({ user: ADMIN_USER, pwd: ADMIN_PASS,
        roles: [{role: "userAdminAnyDatabase", db: ADMIN_DB},
        { role: "readWriteAnyDatabase", db: ADMIN_DB}]
        });

    //Create the Root user named Super
    authDB.createUser({ user: SUPER_USER, pwd: SUPER_PASS,
        roles: [{role: "root", db: ADMIN_DB}]
        });

    //Create the Backup user named guard
    authDB.createUser({ user: GUARD_USER, pwd: GUARD_PASS,
        roles: [{role: "backup", db: ADMIN_DB}, {role: "restore", db: ADMIN_DB}]
        });
} 

//Authenticate first
var m = new Mongo()
var authDB = m.getDB(ADMIN_DB)
authDB.auth(ADMIN_USER, ADMIN_PASS);

// Create a new Database
var db = m.getDB("ogree"+DB_NAME)

//Create a default user
db.createCollection('account');
db.account.insertOne({email: "admin", password: "admin", roles: {"*": "manager"}})

// Create API User
db.createUser({ user: "ogree"+DB_NAME+"Admin", pwd: CUSTOMER_API_PASSWORD,
                roles: [{role: "readWrite", db: "ogree"+DB_NAME}, {role: "restore", db: "admin"}]
                })
