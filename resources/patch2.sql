CREATE TABLE public.subdevice (
    id serial NOT NULL,
    subdevice_name text,
    subdevice_parent_id int NOT NULL,
    subdevice_domain text,
    subdevice_description text,
    PRIMARY KEY (id)
);


CREATE TABLE public.subdevice_attributes (
    id int NOT NULL,
    subdevice_pos_x_y text,
    subdevice_pos_x_y_unit text,
    subdevice_pos_z text,
    subdevice_pos_z_unit text,
    subdevice_template text,
    subdevice_orientation text,
    subdevice_size text,
    subdevice_size_unit text,
    subdevice_height text,
    subdevice_height_unit text,
    subdevice_type text, 
    subdevice_vendor text,
    subdevice_model text,
    subdevice_serial text,
    subdevice_sizeu text,
    subdevice_posu text,
    subdevice_slot text,
    PRIMARY KEY (id)
);

ALTER TABLE public.subdevice ADD CONSTRAINT FK_subdevice__subdevice_parent_id FOREIGN KEY (subdevice_parent_id) REFERENCES public.device(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.subdevice_attributes ADD CONSTRAINT FK_subdevice_attributes__id FOREIGN KEY (id) REFERENCES public.subdevice(id) ON DELETE CASCADE ON UPDATE CASCADE;


ALTER TABLE subdevice ADD CONSTRAINT chk_subdevice_parent CHECK (subdevice_parent_id IS NOT NULL);
ALTER TABLE subdevice_attributes ADD CONSTRAINT chk_subdevice_orientation CHECK (subdevice_orientation IN ('front', 'rear', 'front flipped', 'rear flipped'));
--ALTER TABLE device_attributes ADD CONSTRAINT chk_device_pxyu CHECK (device_pos_x_y_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
--ALTER TABLE device_attributes ADD CONSTRAINT chk_device_pzu CHECK (device_pos_z_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE subdevice_attributes ADD CONSTRAINT chk_subdevice_sizeu CHECK (subdevice_size_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE subdevice_attributes ADD CONSTRAINT chk_subdevice_phu CHECK (subdevice_height_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));




CREATE TABLE public.subdevice1 (
    id serial NOT NULL,
    subdevice1_name text,
    subdevice1_parent_id int NOT NULL,
    subdevice1_domain text,
    subdevice1_description text,
    PRIMARY KEY (id)
);


CREATE TABLE public.subdevice1_attributes (
    id int NOT NULL,
    subdevice1_pos_x_y text,
    subdevice1_pos_x_y_unit text,
    subdevice1_pos_z text,
    subdevice1_pos_z_unit text,
    subdevice1_template text,
    subdevice1_orientation text,
    subdevice1_size text,
    subdevice1_size_unit text,
    subdevice1_height text,
    subdevice1_height_unit text,
    subdevice1_type text, 
    subdevice1_vendor text,
    subdevice1_model text,
    subdevice1_serial text,
    subdevice1_sizeu text,
    subdevice1_posu text,
    subdevice1_slot text,
    PRIMARY KEY (id)
);

ALTER TABLE public.subdevice1 ADD CONSTRAINT FK_subdevice1__subdevice1_parent_id FOREIGN KEY (subdevice1_parent_id) REFERENCES public.subdevice(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.subdevice1_attributes ADD CONSTRAINT FK_subdevice1_attributes__id FOREIGN KEY (id) REFERENCES public.subdevice1(id) ON DELETE CASCADE ON UPDATE CASCADE;


ALTER TABLE subdevice1 ADD CONSTRAINT chk_subdevice1_parent CHECK (subdevice1_parent_id IS NOT NULL);
ALTER TABLE subdevice1_attributes ADD CONSTRAINT chk_subdevice1_orientation CHECK (subdevice1_orientation IN ('front', 'rear', 'front flipped', 'rear flipped'));
--ALTER TABLE device_attributes ADD CONSTRAINT chk_device_pxyu CHECK (device_pos_x_y_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
--ALTER TABLE device_attributes ADD CONSTRAINT chk_device_pzu CHECK (device_pos_z_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE subdevice1_attributes ADD CONSTRAINT chk_subdevice1_sizeu CHECK (subdevice1_size_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));
ALTER TABLE subdevice1_attributes ADD CONSTRAINT chk_subdevice1_phu CHECK (subdevice1_height_unit IN ('mm', 'cm', 'm', 'U', 'OU', 'tile'));