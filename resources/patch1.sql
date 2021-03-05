
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

--SET enable_experimental_alter_column_type_general = true;
-- alter table device alter column device_description SET DATA TYPE text;
-- alter table rack alter column device_description SET DATA TYPE text;
-- alter table room alter column device_description SET DATA TYPE text;
-- alter table building alter column device_description SET DATA TYPE text;
-- alter table site alter column device_description SET DATA TYPE text;
-- alter table tenant alter column device_description SET DATA TYPE text;
--SET enable_experimental_alter_column_type_general = false;



