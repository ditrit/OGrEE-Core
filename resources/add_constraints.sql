
-- ALTER TABLE Tenant ADD CONSTRAINT chk_tenant_parent CHECK (tenant_parent_id IS NULL);

ALTER TABLE Site ADD CONSTRAINT chk_site_parent CHECK (site_parent_id IS NOT NULL);

ALTER TABLE site_attributes ADD CONSTRAINT chk_site_orientation CHECK (site_orientation IN ('EN', 'NW', 'WS', 'SE'));

ALTER TABLE Building ADD CONSTRAINT chk_bldg_parent CHECK (bldg_parent_id IS NOT NULL);
ALTER TABLE building_attributes ADD CONSTRAINT chk_bldg_pxyu CHECK (bldg_pos_x_y_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE building_attributes ADD CONSTRAINT chk_bldg_pzu CHECK (bldg_pos_z_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE building_attributes ADD CONSTRAINT chk_bldg_sizeu CHECK (bldg_size_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE building_attributes ADD CONSTRAINT chk_bldg_phu CHECK (bldg_height_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));



ALTER TABLE Room ADD CONSTRAINT chk_room_parent CHECK (room_parent_id IS NOT NULL);
ALTER TABLE room_attributes ADD CONSTRAINT chk_room_orientation CHECK (room_orientation IN ('-E-N', '-E+N', '+E-N', '+E+N','-N-W', '-N+W', '+N-W', '+N+W','-W-S', '-W+S', '+W-S', '+W+S', '-S-E', '-S+E', '+S-E', '+S+E'));
ALTER TABLE room_attributes ADD CONSTRAINT chk_room_pxyu CHECK (room_pos_x_y_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE room_attributes ADD CONSTRAINT chk_room_pzu CHECK (room_pos_z_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE room_attributes ADD CONSTRAINT chk_room_sizeu CHECK (room_size_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE room_attributes ADD CONSTRAINT chk_room_phu CHECK (room_height_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));



ALTER TABLE Rack ADD CONSTRAINT chk_rack_parent CHECK (rack_parent_id IS NOT NULL);
ALTER TABLE rack_attributes ADD CONSTRAINT chk_rack_orientation CHECK (rack_orientation IN ('front', 'rear', 'left', 'right'));
ALTER TABLE rack_attributes ADD CONSTRAINT chk_rack_pxyu CHECK (rack_pos_x_y_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
--ALTER TABLE rack_attributes ADD CONSTRAINT chk_rack_pzu CHECK (rack_pos_z_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE rack_attributes ADD CONSTRAINT chk_rack_sizeu CHECK (rack_size_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE rack_attributes ADD CONSTRAINT chk_rack_phu CHECK (rack_height_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));



ALTER TABLE Device ADD CONSTRAINT chk_device_parent CHECK (device_parent_id IS NOT NULL);
ALTER TABLE device_attributes ADD CONSTRAINT chk_device_orientation CHECK (device_orientation IN ('front', 'rear', 'front flipped', 'rear flipped'));
ALTER TABLE device_attributes ADD CONSTRAINT chk_device_pxyu CHECK (device_pos_x_y_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
--ALTER TABLE device_attributes ADD CONSTRAINT chk_device_pzu CHECK (device_pos_z_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE device_attributes ADD CONSTRAINT chk_device_sizeu CHECK (device_size_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE device_attributes ADD CONSTRAINT chk_device_phu CHECK (device_height_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
