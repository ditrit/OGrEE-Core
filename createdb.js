/////
// NOTE - WARNING
// This code is mirrored in the API found in models.base.go:CreateTenantDB
// If you modify this script be sure to make the same changes in said
// function!
//////

//localhost = 127.0.0.1


//Check if dbName was passed as argument
//Otherwise use "ogree"
try {
  dbName;
} catch(e) {
  dbName = "develop"
}

var m = new Mongo()
var db = m.getDB(dbName)

db.createCollection('account');
db.createCollection('tenant');
db.createCollection('site');
db.createCollection('building');
db.createCollection('room');
db.createCollection('rack');
db.createCollection('device');

//Template Collections
db.createCollection('room_template');
db.createCollection('obj_template');

//Group Collections
db.createCollection('group');


//Nonhierarchal objects
db.createCollection('ac');
db.createCollection('panel');
db.createCollection('separator');
db.createCollection('row');
db.createCollection('tile');
db.createCollection('cabinet');
db.createCollection('corridor');
db.createCollection('sensor');

//Stray Objects
db.createCollection('stray_device');


//Enfore unique Tenant Names
db.tenant.createIndex( {"name":1}, { unique: true } );

//Enforce unique children
db.site.createIndex({parentId:1, name:1}, { unique: true });
db.building.createIndex({parentId:1, name:1}, { unique: true });
db.room.createIndex({parentId:1, name:1}, { unique: true });
db.rack.createIndex({parentId:1, name:1}, { unique: true });
db.device.createIndex({parentId:1, name:1}, { unique: true });
//Enforcing that the Parent Exists is done at the ORM Level for now


//Make slugs unique identifiers for templates
db.room_template.createIndex({slug:1}, { unique: true });
db.obj_template.createIndex({slug:1}, { unique: true });


//Unique children restriction for nonhierarchal objects and sensors
db.ac.createIndex({parentId:1, name:1}, { unique: true });
db.panel.createIndex({parentId:1, name:1}, { unique: true });
db.separator.createIndex({parentId:1, name:1}, { unique: true });
db.row.createIndex({parentId:1, name:1}, { unique: true });
db.tile.createIndex({parentId:1, name:1}, { unique: true });
db.cabinet.createIndex({parentId:1, name:1}, { unique: true });
db.corridor.createIndex({parentId:1, name:1}, { unique: true });

//Enforce unique children sensors
db.sensor.createIndex({parentId:1, type:1, name:1}, { unique: true });

//Enforce unique Group names 
db.group.createIndex({parentId:1, name:1}, { unique: true });

//Enforce unique stray objects
db.stray_device.createIndex({parentId:1, name:1}, { unique: true });


//Update customer record table
var odb = m.getDB("ogree")
odb.customer.insertOne({"name": dbName});