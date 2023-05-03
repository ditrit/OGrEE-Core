/////
// NOTE
// This creates a DB that maintains a list of customer DBs
// with a customer collection
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
// CONSTANT DECLARATIONS START
//
DB_NAME;
CUSTOMER_API_PASS;
CUSTOMER_RECORDS_DB;

ADMIN_DB;
SUPER_USER;
SUPER_PASS;

ADMIN_USER;
ADMIN_PASS;

GUARD_USER;
GUARD_PASS;

//
// CONSTANT DECLARATIONS END
//




var m = new Mongo()
var authDB = m.getDB(ADMIN_DB)

//Create the Root user named Super
db.createUser({ user: SUPER_USER, pwd: SUPER_PASS,
                roles: [{role: "root", db: ADMIN_DB}]
                })

//Create the Admin user
db.createUser({ user: ADMIN_USER, pwd: ADMIN_PASS,
                roles: [{role: "userAdminAnyDatabase", db: ADMIN_DB},
                { role: "readWriteAnyDatabase", db: ADMIN_DB}]
                })

//Create the Backup user named guard
db.createUser({ user: GUARD_USER, pwd: GUARD_PASS,
                roles: [{role: "backup", db: ADMIN_DB}, {role: "restore", db: ADMIN_DB}]
                })

//Create customer record collection                
var db = m.getDB(CUSTOMER_RECORDS_DB);
db.createCollection('customer');
db.customer.createIndex({name:1}, { unique: true });





/////
// Create a new Database
//////

//Authenticate first
var m = new Mongo()
var authDB = m.getDB("test")
authDB.auth(ADMIN_USER, ADMIN_PASS);



//First Update customer record collection
var odb = m.getDB(CUSTOMER_RECORDS_DB)
odb.customer.insertOne({"name": DB_NAME});


//Then Create the customer DB
var db = m.getDB("ogree"+DB_NAME)
db.createCollection('account');
db.account.insertOne( { email: "admin", password: "admin", roles: {"*":"manager"} } );
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
db.createCollection('sensor');

//Stray Objects
db.createCollection('stray_device');
db.createCollection('stray_sensor');


//Enfore unique Tenant Names
db.domain.createIndex( {parentId:1, name:1}, { unique: true } );

//Enforce unique children
db.site.createIndex({name:1}, { unique: true });
db.building.createIndex({parentId:1, name:1}, { unique: true });
db.room.createIndex({parentId:1, name:1}, { unique: true });
db.rack.createIndex({parentId:1, name:1}, { unique: true });
db.device.createIndex({parentId:1, name:1}, { unique: true });
//Enforcing that the Parent Exists is done at the ORM Level for now


//Make slugs unique identifiers for templates
db.room_template.createIndex({slug:1}, { unique: true });
db.obj_template.createIndex({slug:1}, { unique: true });
db.bldg_template.createIndex({slug:1}, { unique: true });


//Unique children restriction for nonhierarchal objects and sensors
db.ac.createIndex({parentId:1, name:1}, { unique: true });
db.panel.createIndex({parentId:1, name:1}, { unique: true });
db.cabinet.createIndex({parentId:1, name:1}, { unique: true });
db.corridor.createIndex({parentId:1, name:1}, { unique: true });

//Enforce unique children sensors
db.sensor.createIndex({parentId:1, type:1, name:1}, { unique: true });

//Enforce unique Group names 
db.group.createIndex({parentId:1, name:1}, { unique: true });

//Enforce unique stray objects
db.stray_device.createIndex({parentId:1,name:1}, { unique: true });
db.stray_sensor.createIndex({name:1}, { unique: true });


// Create Respective API User
// To create a new customer you should access the 
// running container, run the createdb.js and createUser.js scripts
// contained in the home folder 

//Authenticate first
var m = new Mongo()
var authDB = m.getDB("test")
authDB.auth(ADMIN_USER, ADMIN_PASS);


db.createUser({ user: "ogree"+DB_NAME+"Admin", pwd: CUSTOMER_API_PASS,
                roles: [{role: "readWrite", db: "ogree"+DB_NAME}]
                })
