CREATE TABLE public.tenant (
    tenant_id text NOT NULL,
    tenant_name text,
    tenant_parent_id text,
    tenant_domain text,
    tenant_description text[],
    PRIMARY KEY (tenant_id)
);


CREATE TABLE public.site (
    site_id text NOT NULL,
    site_name text,
    site_parent_id text NOT NULL,
    site_domain text,
    site_description text[],
    PRIMARY KEY (site_id)
);

CREATE INDEX ON public.site
    (site_parent_id);


CREATE TABLE public.building (
    bldg_id text NOT NULL,
    bldg_name text,
    bldg_parent_id text NOT NULL,
    bldg_domain text,
    bldg_description text[],
    PRIMARY KEY (bldg_id)
);

CREATE INDEX ON public.building
    (bldg_parent_id);


CREATE TABLE public.room (
    room_id text NOT NULL,
    room_name text,
    room_parent_id text NOT NULL,
    room_domain text,
    room_description text[],
    PRIMARY KEY (room_id)
);

CREATE INDEX ON public.room
    (room_parent_id);


CREATE TABLE public.rack (
    rack_id text NOT NULL,
    rack_name text,
    rack_parent_id text NOT NULL,
    rack_domain text,
    rack_description text[],
    PRIMARY KEY (rack_id)
);

CREATE INDEX ON public.rack
    (rack_parent_id);


CREATE TABLE public.device (
    device_id text NOT NULL,
    device_name text,
    device_parent_id text NOT NULL,
    device_domain text,
    device_description text[],
    PRIMARY KEY (device_id)
);

CREATE INDEX ON public.device
    (device_parent_id);


CREATE TABLE public.size (
    name text NOT NULL,
    non_standard_attr_length text NOT NULL,
    PRIMARY KEY (name)
);


CREATE TABLE public.tenant_attributes (
    tenant_id text NOT NULL,
    tenant_color text,
    main_contact text,
    main_phone text,
    main_email text,
    PRIMARY KEY (tenant_id)
);


CREATE TABLE public.site_attributes (
    site_id text NOT NULL,
    site_orientation text,
    usable_color text,
    reserved_color text,
    technical_color text,
    address text,
    zipcode text,
    city text,
    country text,
    gps text,
    PRIMARY KEY (site_id)
);


CREATE TABLE public.building_attributes (
    bldg_id text NOT NULL,
    bldg_pos_x_y text,
    bldg_pos_x_y_unit text,
    bldg_pos_z text,
    bldg_pos_z_unit text,
    bldg_size text,
    bldg_size_unit text,
    bldg_height text,
    bldg_height_unit text,
    bldg_nb_floors text,
    PRIMARY KEY (bldg_id)
);


CREATE TABLE public.room_attributes (
    room_id text NOT NULL,
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
    PRIMARY KEY (room_id)
);


CREATE TABLE public.rack_attributes (
    rack_id text NOT NULL,
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
    PRIMARY KEY (rack_id)
);


CREATE TABLE public.device_attributes (
    device_id text NOT NULL,
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
    PRIMARY KEY (device_id)
);


ALTER TABLE public.site ADD CONSTRAINT FK_site__site_parent_id FOREIGN KEY (site_parent_id) REFERENCES public.tenant(tenant_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.building ADD CONSTRAINT FK_building__bldg_parent_id FOREIGN KEY (bldg_parent_id) REFERENCES public.site(site_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.room ADD CONSTRAINT FK_room__room_parent_id FOREIGN KEY (room_parent_id) REFERENCES public.building(bldg_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.rack ADD CONSTRAINT FK_rack__rack_parent_id FOREIGN KEY (rack_parent_id) REFERENCES public.room(room_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.device ADD CONSTRAINT FK_device__device_parent_id FOREIGN KEY (device_parent_id) REFERENCES public.rack(rack_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.tenant_attributes ADD CONSTRAINT FK_tenant_attributes__tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenant(tenant_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.site_attributes ADD CONSTRAINT FK_site_attributes__site_id FOREIGN KEY (site_id) REFERENCES public.site(site_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.building_attributes ADD CONSTRAINT FK_building_attributes__bldg_id FOREIGN KEY (bldg_id) REFERENCES public.building(bldg_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.room_attributes ADD CONSTRAINT FK_room_attributes__room_id FOREIGN KEY (room_id) REFERENCES public.room(room_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.rack_attributes ADD CONSTRAINT FK_rack_attributes__rack_id FOREIGN KEY (rack_id) REFERENCES public.rack(rack_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE public.device_attributes ADD CONSTRAINT FK_device_attributes__device_id FOREIGN KEY (device_id) REFERENCES public.device(device_id) ON DELETE CASCADE ON UPDATE CASCADE;
