--- TENANT ---
-- tenant_id | tenant_name | tenant_parent_id | tenant_domain | tenant_description | 
-- tenant_color | main_contact | main_phone | main_email

INSERT INTO tenant (tenant_id, tenant_name, tenant_domain, tenant_description)
VALUES ('1','CEA','Reseau', '{Cant evaluate anything}');

INSERT INTO tenant_attributes (tenant_id, tenant_color, main_contact, main_phone, main_email)
VALUES ('1','Red', 'Visit them', '000', 'cea@cea.com');

INSERT INTO tenant (tenant_id, tenant_name, tenant_domain, tenant_description) 
VALUES ('2','Mitsubishi','Reseau', '{TriAge}');

INSERT INTO tenant_attributes (tenant_id, tenant_color, main_contact, main_phone, main_email)
VALUES ('2', 'Red', 'website', '000', 'mitsubishi.com');

INSERT INTO tenant (tenant_id, tenant_name) VALUES ('3','SANYO');
INSERT INTO tenant_attributes (tenant_id) VALUES ('3');


--- SITE ---
/*
site_id | site_name | site_parent_id | site_domain 
| site_description 

| site_orientation | usable_color 
| reserved_color | technical_color | address | zipcode 
| city | country | gps
*/

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('0', 'siteA', '1');
INSERT INTO site_attributes (site_id) VALUES ('0');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('1', 'siteB', '1');
INSERT INTO site_attributes (site_id) VALUES ('1');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('2', 'siteC', '1');
INSERT INTO site_attributes (site_id) VALUES ('2');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('3', 'siteA', '2');
INSERT INTO site_attributes (site_id) VALUES ('3');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('4', 'siteB', '2');
INSERT INTO site_attributes (site_id) VALUES ('4');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('5', 'siteC', '2');
INSERT INTO site_attributes (site_id) VALUES ('5');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('6', 'siteA', '3');
INSERT INTO site_attributes (site_id) VALUES ('6');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('7', 'siteB', '3');
INSERT INTO site_attributes (site_id) VALUES ('7');

INSERT INTO site (site_id, site_name, site_parent_id) VALUES ('8', 'siteC', '3');
INSERT INTO site_attributes (site_id) VALUES ('8');



--- BUILDING --- 
/*bldg_id | bldg_name | 
bldg_parent_id | bldg_domain | 
bldg_description 

| bldg_pos_x_y | 
bldg_pos_x_y_unit | bldg_pos_z | 
bldg_pos_z_unit | bldg_size | bldg_size_unit | 
bldg_height | bldg_height_unit | bldg_nb_floors
*/
INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('0', 'bldgA', '0');
INSERT INTO building_attributes (bldg_id) VALUES ('0');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('0', 'bldgA', '1');
INSERT INTO building_attributes (bldg_id) VALUES ('0');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('1', 'bldgB', '1');
INSERT INTO building_attributes (bldg_id) VALUES ('1');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('2', 'bldgC', '1');
INSERT INTO building_attributes (bldg_id) VALUES ('2');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('3', 'bldgA', '2');
INSERT INTO building_attributes (bldg_id) VALUES ('3');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('4', 'bldgB', '2');
INSERT INTO building_attributes (bldg_id) VALUES ('4');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('5', 'bldgC', '2');
INSERT INTO building_attributes (bldg_id) VALUES ('5');


INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('6', 'bldgA', '3');
INSERT INTO building_attributes (bldg_id) VALUES ('6');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('7', 'bldgB', '3');
INSERT INTO building_attributes (bldg_id) VALUES ('7');

INSERT INTO building (bldg_id, bldg_name, bldg_parent_id) VALUES ('8', 'bldgC', '3');
INSERT INTO building_attributes (bldg_id) VALUES ('8');



--- ROOM ---
/*
room_id | room_name | room_parent_id 
| room_domain | room_description 

| room_pos_x_y 
| room_pos_x_y_unit | room_pos_z | room_pos_z_unit 
| room_template | room_orientation | room_size 
| room_size_unit | room_height | room_height_unit
*/
INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('0', 'roomA', '1');
INSERT INTO room_attributes (room_id) VALUES ('0');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('1', 'roomB', '1');
INSERT INTO room_attributes (room_id) VALUES ('1');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('2', 'roomC', '1');
INSERT INTO room_attributes (room_id) VALUES ('2');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('3', 'roomA', '2');
INSERT INTO room_attributes (room_id) VALUES ('3');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('4', 'roomB', '2');
INSERT INTO room_attributes (room_id) VALUES ('4');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('5', 'roomC', '2');
INSERT INTO room_attributes (room_id) VALUES ('5');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('6', 'roomA', '3');
INSERT INTO room_attributes (room_id) VALUES ('6');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('7', 'roomB', '3');
INSERT INTO room_attributes (room_id) VALUES ('7');

INSERT INTO room (room_id, room_name, room_parent_id) VALUES ('8', 'roomC', '3');
INSERT INTO room_attributes (room_id) VALUES ('8');


--- RACK --- 
/*
rack_id | rack_name | rack_parent_id 
| rack_domain | rack_description 

| rack_pos_x_y | rack_pos_x_y_unit 
| rack_pos_z | rack_pos_z_unit | rack_template 
| rack_orientation | rack_size | rack_size_unit 
| rack_height | rack_height_unit
*/

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('0', 'rackA', '1');
INSERT INTO rack_attributes (rack_id) VALUES ('0');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('1', 'rackB', '1');
INSERT INTO rack_attributes (rack_id) VALUES ('1');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('2', 'rackC', '1');
INSERT INTO rack_attributes (rack_id) VALUES ('2');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('3', 'rackA', '2');
INSERT INTO rack_attributes (rack_id) VALUES ('3');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('4', 'rackB', '2');
INSERT INTO rack_attributes (rack_id) VALUES ('4');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('5', 'rackC', '2');
INSERT INTO rack_attributes (rack_id) VALUES ('5');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('6', 'rackA', '3');
INSERT INTO rack_attributes (rack_id) VALUES ('6');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('7', 'rackB', '3');
INSERT INTO rack_attributes (rack_id) VALUES ('7');

INSERT INTO rack (rack_id, rack_name, rack_parent_id) VALUES ('8', 'rackC', '3');
INSERT INTO rack_attributes (rack_id) VALUES ('8');



--- DEVICE ---
/*
device_id | device_name | device_parent_id 
| device_domain | device_description 

| device_pos_x_y | device_pos_x_y_unit 
| device_pos_z | device_pos_z_unit | device_template 
| device_orientation | device_size | device_size_unit 
| device_height | device_height_unit
*/
INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('0', 'deviceA', '1');
INSERT INTO device_attributes (device_id) VALUES ('0');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('1', 'deviceB', '1');
INSERT INTO device_attributes (device_id) VALUES ('1');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('2', 'deviceC', '1');
INSERT INTO device_attributes (device_id) VALUES ('2');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('3', 'deviceA', '2');
INSERT INTO device_attributes (device_id) VALUES ('3');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('4', 'deviceB', '2');
INSERT INTO device_attributes (device_id) VALUES ('4');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('5', 'deviceC', '2');
INSERT INTO device_attributes (device_id) VALUES ('5');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('6', 'deviceA', '3');
INSERT INTO device_attributes (device_id) VALUES ('6');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('7', 'deviceB', '3');
INSERT INTO device_attributes (device_id) VALUES ('7');

INSERT INTO device (device_id, device_name, device_parent_id) VALUES ('8', 'deviceC', '3');
INSERT INTO device_attributes (device_id) VALUES ('8');



--- SIZE ---
--- Side Table Joint
INSERT INTO size (name, non_standard_attr_length) VALUES ('tenant', '4');
INSERT INTO size (name, non_standard_attr_length) VALUES ('site', '9');
INSERT INTO size (name, non_standard_attr_length) VALUES ('building', '9');
INSERT INTO size (name, non_standard_attr_length) VALUES ('room', '10');
INSERT INTO size (name, non_standard_attr_length) VALUES ('rack', '10');
INSERT INTO size (name, non_standard_attr_length) VALUES ('device', '10');
