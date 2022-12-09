/////
// NOTE
// This initialises a DB that maintains a list of customer DBs
// via the customer collection and creates an admin user
// MongoDB docker image will execute scripts in alphabetical order
//////

var m = new Mongo()
var authDB = m.getDB("admin")

//Create the Root user named Super
db.createUser({ user: "super", pwd: "superpassword",
                roles: [{role: "root", db: "admin"}]
                })

//Create the Admin user
db.createUser({ user: "admin", pwd: "adminpassword",
                roles: [{role: "userAdminAnyDatabase", db: "admin"},
                { role: "readWriteAnyDatabase", db: "admin"}]
                })

//Create the Backup user named guard
db.createUser({ user: "guard", pwd: "adminpassword",
                roles: [{role: "backup", db: "admin"}, {role: "restore", db: "admin"}]
                })

var db = m.getDB("ogree");
db.createCollection('customer');
db.customer.createIndex({name:1}, { unique: true });