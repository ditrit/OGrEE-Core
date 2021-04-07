
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
SET sql_safe_updates = false;
ALTER TABLE device DROP column device_description;
ALTER TABLE rack DROP column rack_description;
ALTER TABLE room DROP column room_description;
ALTER TABLE building DROP column bldg_description;
ALTER TABLE site DROP column site_description;
ALTER TABLE tenant DROP column tenant_description;

ALTER TABLE device ADD column device_description text;
ALTER TABLE rack ADD column rack_description text;
ALTER TABLE room ADD column room_description text;
ALTER TABLE building ADD column bldg_description text;
ALTER TABLE site ADD column site_description text;
ALTER TABLE tenant ADD column tenant_description text;
SET sql_safe_updates = true;
SET enable_experimental_alter_column_type_general = false;

ALTER TABLE account ADD column created_at timestamp with time zone;
ALTER TABLE account ADD column deleted_at timestamp with time zone;
ALTER TABLE account ADD column updated_at timestamp with time zone;

CREATE INDEX idx_account_deleted_at ON "account"(deleted_at);


ALTER TABLE room ADD column technical text;
ALTER TABLE room ADD column reserved text;