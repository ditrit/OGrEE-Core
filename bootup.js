/////
// NOTE
// This initialises a DB that maintains a list of customer DBs
// via the customer collection
// MongoDB docker image will execute scripts in alphabetical order
//////

var m = new Mongo();
var db = m.getDB("ogree");
db.createCollection('customer');
db.customer.createIndex({name:1}, { unique: true });