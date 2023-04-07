
/////
// Create a new customer
//////

//
// CONSTANT DECLARATIONS START
//
// They should already be initialised by 
// the container orchestration environment
//
// DB_NAME and CUSTOMER_API_PASS should be provided by the 
// corresponding 'addCustomer.sh' script 
//
DB_NAME;
CUSTOMER_RECORDS_DB;
CUSTOMER_API_PASS;

ADMIN_USER;
ADMIN_PASS;
//
// CONSTANT DECLARATIONS END
//


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
