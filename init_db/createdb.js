//localhost = 127.0.0.1
//
var m = new Mongo()
var db = m.getDB("ogree")

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