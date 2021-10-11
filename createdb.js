//localhost = 127.0.0.1
var db = connect('localhost:27017/ogree')

db.createCollection('account');
db.createCollection('tenant');
db.createCollection('site');
db.createCollection('building');
db.createCollection('room');
db.createCollection('rack');
db.createCollection('device');
db.createCollection('subdevice');
db.createCollection('subdevice1');

//Template Collections
db.createCollection('room_template');
db.createCollection('rack_template');
db.createCollection('device_template');


//Enfore unique Tenant Names
db.tenant.createIndex( {"name":1}, { unique: true } );

//Enforce unique children
db.site.createIndex({parentId:1, name:1}, { unique: true });
db.building.createIndex({parentId:1, name:1}, { unique: true });
db.room.createIndex({parentId:1, name:1}, { unique: true });
db.rack.createIndex({parentId:1, name:1}, { unique: true });
db.device.createIndex({parentId:1, name:1}, { unique: true });
db.subdevice.createIndex({parentId:1, name:1}, { unique: true });
db.subdevice1.createIndex({parentId:1, name:1}, { unique: true });

//Enforcing that the Parent Exists is done at the ORM Level for now


//Make slugs as unique identifiers for templates
db.room_template.createIndex({slug:1}, { unique: true });
db.rack_template.createIndex({slug:1}, { unique: true });
db.device_template.createIndex({slug:1}, { unique: true });