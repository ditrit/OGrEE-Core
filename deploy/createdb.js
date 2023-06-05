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
db.createCollection('account');
db.createCollection('domain');
db.createCollection('site');
db.createCollection('building');
db.createCollection('room');
db.createCollection('rack');
db.createCollection('device');

//Template Collections
db.createCollection('room_template');
db.createCollection('obj_template');
db.createCollection('bldg_template');

//Group Collections
db.createCollection('group');

//Nonhierarchal objects
db.createCollection('ac');
db.createCollection('panel');
db.createCollection('cabinet');
db.createCollection('corridor');

//Stray Objects
db.createCollection('stray_device');

//Enforce unique children
db.domain.createIndex( {parentId:1, name:1}, { unique: true } );
db.site.createIndex({name:1}, { unique: true });
db.building.createIndex({parentId:1, name:1}, { unique: true });
db.room.createIndex({parentId:1, name:1}, { unique: true });
db.rack.createIndex({parentId:1, name:1}, { unique: true });
db.device.createIndex({parentId:1, name:1}, { unique: true });

//Make slugs unique identifiers for templates
db.room_template.createIndex({slug:1}, { unique: true });
db.obj_template.createIndex({slug:1}, { unique: true });
db.bldg_template.createIndex({slug:1}, { unique: true });

//Unique children restriction for nonhierarchal objects and sensors
db.ac.createIndex({parentId:1, name:1}, { unique: true });
db.panel.createIndex({parentId:1, name:1}, { unique: true });
db.cabinet.createIndex({parentId:1, name:1}, { unique: true });
db.corridor.createIndex({parentId:1, name:1}, { unique: true });

//Enforce unique Group names 
db.group.createIndex({parentId:1, name:1}, { unique: true });

//Enforce unique stray objects
db.stray_device.createIndex({parentId:1,name:1}, { unique: true });

//Create a default domain and user
db.domain.insertOne({name: DB_NAME, hierarchyName: DB_NAME, category: "domain", 
    attributes:{color:"ffffff"}, description:[], createdData: new Date(), lastUpdated: new Date()})
db.account.insertOne({email: "admin", password: "admin", roles: {"*": "manager"}})

// Create API User
db.createUser({ user: "ogree"+DB_NAME+"Admin", pwd: CUSTOMER_API_PASSWORD,
                roles: [{role: "readWrite", db: "ogree"+DB_NAME}]
                })
