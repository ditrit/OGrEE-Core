/////
// NOTE
// This initialises a DB that maintains a list of customer DBs
// via the customer collection
// MongoDB docker image will execute scripts in alphabetical order
//////

var m = new Mongo()
var authDB = m.getDB("admin")
db.createUser({ user: "admin", pwd: "adminpassword",
                roles: [{role: "userAdminAnyDatabase", db: "admin"},
                { role: "readWriteAnyDatabase", db: "admin"}]
                })
//db.createUser({ user: "admin", pwd: "adminpassword", 
//                roles: [{ role: "userAdminAnyDatabase", db: "admin" }] })

var db = m.getDB("ogree");
db.createCollection('customer');
db.customer.createIndex({name:1}, { unique: true });

  //{ role: "userAdminAnyDatabase", db: "admin" },