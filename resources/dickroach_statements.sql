CREATE TABLE "accounts" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"email" text,"password" text,"token" text , PRIMARY KEY ("id"))  
CREATE INDEX idx_accounts_deleted_at ON "accounts"(deleted_at)
CREATE TABLE "tenants" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"name" text,"category" text,"desc" text,"domain" text,"color" text , PRIMARY KEY ("id"))
CREATE INDEX idx_tenants_deleted_at ON "tenants"(deleted_at)
CREATE TABLE "sites" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"name" text,"category" text,"desc" text,"domain" integer,"color" text,"orientation" text , PRIMARY KEY ("id"))
CREATE INDEX idx_sites_deleted_at ON "sites"(deleted_at)
CREATE TABLE "buildings" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"name" text,"category" text,"desc" text,"domain" integer,"pos_x" numeric,"pos_y" numeric,"pos_u" text,"pos_z" numeric,"pos_zu" text,"size" numeric,"size_u" text,"height" numeric,"height_u" text , PRIMARY KEY ("id"))
CREATE INDEX idx_buildings_deleted_at ON "buildings"(deleted_at)
CREATE TABLE "rooms" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"name" text,"category" text,"desc" text,"domain" integer,"orientation" text,"pos_x" numeric,"pos_y" numeric,"pos_u" text,"pos_z" numeric,"pos_zu" text,"size" numeric,"size_u" text,"height" numeric,"height_u" text , PRIMARY KEY ("id"))
CREATE INDEX idx_rooms_deleted_at ON "rooms"(deleted_at)
CREATE TABLE "racks" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"name" text,"category" text,"desc" text,"domain" integer,"color" text,"orientation" text , PRIMARY KEY ("id"))
CREATE INDEX idx_racks_deleted_at ON "racks"(deleted_at)
CREATE TABLE "devices" ("id" serial,"created_at" timestamp with time zone,"updated_at" timestamp with time zone,"deleted_at" timestamp with time zone,"name" text,"category" text,"desc" text,"domain" integer,"color" text,"orientation" text , PRIMARY KEY ("id"))
CREATE INDEX idx_devices_deleted_at ON "devices"(deleted_at)


CREATE TABLE public.accounts (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  email STRING NULL,
  password STRING NULL,
  token STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  INDEX idx_accounts_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, email, password, token)
)


CREATE TABLE public.tenants (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  name STRING NULL,
  category STRING NULL,
  "desc" STRING NULL,
  domain STRING NULL,
  color STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  INDEX idx_tenants_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, name, category, "desc", domain, color)
)


CREATE TABLE public.sites (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  name STRING NULL,
  category STRING NULL,
  "desc" STRING NULL,
  domain INT8 NULL,
  color STRING NULL,
  orientation STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  CONSTRAINT sites_domain_tenants_id_foreign FOREIGN KEY (domain) REFERENCES public.tenants(id) ON DELETE CASCADE ON UPDATE CASCADE,
  INDEX idx_sites_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, name, category, "desc", domain, color, orientation)
)


CREATE TABLE public.buildings (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  name STRING NULL,
  category STRING NULL,
  "desc" STRING NULL,
  domain INT8 NULL,
  pos_x DECIMAL NULL,
  pos_y DECIMAL NULL,
  pos_u STRING NULL,
  pos_z DECIMAL NULL,
  pos_zu STRING NULL,
  size DECIMAL NULL,
  size_u STRING NULL,
  height DECIMAL NULL,
  height_u STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  CONSTRAINT buildings_domain_sites_id_foreign FOREIGN KEY (domain) REFERENCES public.sites(id) ON DELETE CASCADE ON UPDATE CASCADE,
  INDEX idx_buildings_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, name, category, "desc", domain, pos_x, pos_y, pos_u, pos_z, pos_zu, size, size_u, height, height_u)
)


CREATE TABLE public.rooms (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  name STRING NULL,
  category STRING NULL,
  "desc" STRING NULL,
  domain INT8 NULL,
  orientation STRING NULL,
  pos_x DECIMAL NULL,
  pos_y DECIMAL NULL,
  pos_u STRING NULL,
  pos_z DECIMAL NULL,
  pos_zu STRING NULL,
  size DECIMAL NULL,
  size_u STRING NULL,
  height DECIMAL NULL,
  height_u STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  CONSTRAINT rooms_domain_buildings_id_foreign FOREIGN KEY (domain) REFERENCES public.buildings(id) ON DELETE CASCADE ON UPDATE CASCADE,
  INDEX idx_rooms_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, name, category, "desc", domain, orientation, pos_x, pos_y, pos_u, pos_z, pos_zu, size, size_u, height, height_u)
)

CREATE TABLE public.racks (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  name STRING NULL,
  category STRING NULL,
  "desc" STRING NULL,
  domain INT8 NULL,
  color STRING NULL,
  orientation STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  CONSTRAINT racks_domain_rooms_id_foreign FOREIGN KEY (domain) REFERENCES public.rooms(id) ON DELETE CASCADE ON UPDATE CASCADE,
  INDEX idx_racks_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, name, category, "desc", domain, color, orientation)
)


CREATE TABLE public.devices (
  id INT8 NOT NULL DEFAULT unique_rowid(),
  created_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NULL,
  deleted_at TIMESTAMPTZ NULL,
  name STRING NULL,
  category STRING NULL,
  "desc" STRING NULL,
  domain INT8 NULL,
  color STRING NULL,
  orientation STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC),
  CONSTRAINT devices_domain_racks_id_foreign FOREIGN KEY (domain) REFERENCES public.racks(id) ON DELETE CASCADE ON UPDATE CASCADE,
  INDEX idx_devices_deleted_at (deleted_at ASC),
  FAMILY "primary" (id, created_at, updated_at, deleted_at, name, category, "desc", domain, color, orientation)
)
