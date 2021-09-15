//localhost = 127.0.0.1
var db = connect('localhost:27017/ogree')

db.createCollection('tenant');
db.createCollection('site');
db.createCollection('building');
db.createCollection('room');
db.createCollection('rack');
db.createCollection('device');
db.createCollection('subdevice');
db.createCollection('subdevice1');