CREATE TABLE public.tenant (
    id serial NOT NULL,
    tenant_name text,
    tenant_parent_id int,
    tenant_domain text,
    tenant_description text[],
    PRIMARY KEY (id)
);


CREATE TABLE public.site (
    id serial NOT NULL,
    site_name text,
    site_parent_id int NOT NULL,
    site_domain text,
    site_description text[],
    PRIMARY KEY (id)
);

CREATE INDEX ON public.site
    (site_parent_id);


CREATE TABLE public.building (
    id serial NOT NULL,
    bldg_name text,
    bldg_parent_id int NOT NULL,
    bldg_domain text,
    bldg_description text[],
    PRIMARY KEY (id)
);

CREATE INDEX ON public.building
    (bldg_parent_id);


CREATE TABLE public.room (
    id serial NOT NULL,
    room_name text,
    room_parent_id int NOT NULL,
    room_domain text,
    room_description text[],
    PRIMARY KEY (id)
);

CREATE INDEX ON public.room
    (room_parent_id);


CREATE TABLE public.rack (
    id serial NOT NULL,
    rack_name text,
    rack_parent_id int NOT NULL,
    rack_domain text,
    rack_description text[],
    PRIMARY KEY (id)
);

CREATE INDEX ON public.rack
    (rack_parent_id);


CREATE TABLE public.device (
    id serial NOT NULL,
    device_name text,
    device_parent_id int NOT NULL,
    device_domain text,
    device_description text[],
    PRIMARY KEY (id)
);

CREATE INDEX ON public.device
    (device_parent_id);


CREATE TABLE public.size (
    name text NOT NULL,
    non_standard_attr_length text NOT NULL,
    PRIMARY KEY (name)
);


CREATE TABLE public.tenant_attributes (
    id int NOT NULL,
    tenant_color text,
    main_contact text,
    main_phone text,
    main_email text,
    PRIMARY KEY (id)
);


CREATE TABLE public.site_attributes (
    id int NOT NULL,
    site_orientation text,
    usable_color text,
    reserved_color text,
    technical_color text,
    address text,
    zipcode text,
    city text,
    country text,
    gps text,
    PRIMARY KEY (id)
);


CREATE TABLE public.building_attributes (
    id int NOT NULL,
    bldg_pos_x_y text,
    bldg_pos_x_y_unit text,
    bldg_pos_z text,
    bldg_pos_z_unit text,
    bldg_size text,
    bldg_size_unit text,
    bldg_height text,
    bldg_height_unit text,
    bldg_nb_floors text,
    PRIMARY KEY (id)
);


CREATE TABLE public.room_attributes (
    id int NOT NULL,
    room_pos_x_y text,
    room_pos_x_y_unit text,
    room_pos_z text,
    room_pos_z_unit text,
    room_template text,
    room_orientation text,
    room_size text,
    room_size_unit text,
    room_height text,
    room_height_unit text,
    PRIMARY KEY (id)
);


CREATE TABLE public.rack_attributes (
    id int NOT NULL,
    rack_pos_x_y text,
    rack_pos_x_y_unit text,
    rack_pos_z text,
    rack_pos_z_unit text,
    rack_template text,
    rack_orientation text,
    rack_size text,
    rack_size_unit text,
    rack_height text,
    rack_height_unit text,
    PRIMARY KEY (id)
);


CREATE TABLE public.device_attributes (
    id int NOT NULL,
    device_pos_x_y text,
    device_pos_x_y_unit text,
    device_pos_z text,
    device_pos_z_unit text,
    device_template text,
    device_orientation text,
    device_size text,
    device_size_unit text,
    device_height text,
    device_height_unit text,
    PRIMARY KEY (id)
);


CREATE TABLE public.account (
    id serial NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    token text NOT NULL,
    PRIMARY KEY (id)
);


ALTER TABLE public.site ADD CONSTRAINT FK_site__site_parent_id FOREIGN KEY (site_parent_id) REFERENCES public.tenant(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.building ADD CONSTRAINT FK_building__bldg_parent_id FOREIGN KEY (bldg_parent_id) REFERENCES public.site(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.room ADD CONSTRAINT FK_room__room_parent_id FOREIGN KEY (room_parent_id) REFERENCES public.building(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.rack ADD CONSTRAINT FK_rack__rack_parent_id FOREIGN KEY (rack_parent_id) REFERENCES public.room(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.device ADD CONSTRAINT FK_device__device_parent_id FOREIGN KEY (device_parent_id) REFERENCES public.rack(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.tenant_attributes ADD CONSTRAINT FK_tenant_attributes__id FOREIGN KEY (id) REFERENCES public.tenant(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.site_attributes ADD CONSTRAINT FK_site_attributes__id FOREIGN KEY (id) REFERENCES public.site(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.building_attributes ADD CONSTRAINT FK_building_attributes__id FOREIGN KEY (id) REFERENCES public.building(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.room_attributes ADD CONSTRAINT FK_room_attributes__id FOREIGN KEY (id) REFERENCES public.room(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.rack_attributes ADD CONSTRAINT FK_rack_attributes__id FOREIGN KEY (id) REFERENCES public.rack(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.device_attributes ADD CONSTRAINT FK_device_attributes__id FOREIGN KEY (id) REFERENCES public.device(id) ON DELETE CASCADE ON UPDATE CASCADE;