
ALTER TABLE rack_attributes ADD rack_vendor text;
ALTER TABLE rack_attributes ADD rack_type text;
ALTER TABLE rack_attributes ADD rack_model text;
ALTER TABLE rack_attributes ADD rack_serial text;


ALTER TABLE device_attributes ADD device_vendor text;
ALTER TABLE device_attributes ADD device_type text;
ALTER TABLE device_attributes ADD device_model text;
ALTER TABLE device_attributes ADD device_serial text;


-- Possible option to change description columns
-- to single string

SET enable_experimental_alter_column_type_general = true;
ALTER TABLE device ALTER column device_description SET DATA TYPE text;
ALTER TABLE rack ALTER column rack_description SET DATA TYPE text;
ALTER TABLE room ALTER column room_description SET DATA TYPE text;
ALTER TABLE building ALTER column bldg_description SET DATA TYPE text;
ALTER TABLE site ALTER column site_description SET DATA TYPE text;
ALTER TABLE tenant ALTER column tenant_description SET DATA TYPE text;
SET enable_experimental_alter_column_type_general = false;



