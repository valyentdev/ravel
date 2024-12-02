package schema

const UniqueFleetNameConstraint = "fleets_name_idx"

const initialUp = `
CREATE TABLE  namespaces (
    "name" text primary key,
    "created_at" timestamp not null default timezone('utc', now())
);

CREATE TABLE fleets (
    "id" text primary key,
    "name" text not null,
    "namespace" text not null references namespaces(name) on delete cascade,
    "created_at" timestamp not null default timezone('utc', now()),
	"status" text not null default 'active'
);

CREATE UNIQUE INDEX fleets_name_idx ON  fleets USING btree(namespace, name) WHERE status = 'active';
CREATE INDEX fleets_namespace_idx ON fleets(namespace);
CREATE INDEX fleets_status_idx ON fleets(status);

CREATE TABLE machines (
    "id" text primary key,
    "namespace" text not null references namespaces(name) on delete cascade,
    "fleet_id" text not null references fleets(id) on delete cascade,
    "instance_id" text not null default '',
    "machine_version" text not null default '',
    "node" text not null default '',
    "region" text not null default '',
    "created_at" timestamp not null default timezone('utc', now()),
    "updated_at" timestamp not null default timezone('utc', now()),
    "destroyed_at" timestamp default null
);

CREATE INDEX machines_fleet_id_idx ON machines(fleet_id);
CREATE INDEX machines_namespace_idx ON machines(namespace);

CREATE TABLE machine_versions (
    "id" text not null,
    "namespace" text not null references namespaces(name) on delete cascade,
    "machine_id" text not null references machines(id) on delete cascade,
    "resources" jsonb not null default '{}',
    "config" jsonb not null default '{}',
	PRIMARY KEY ("id", "machine_id")
);

CREATE INDEX machine_versions_machine_id_idx ON machine_versions(machine_id);

CREATE TABLE machine_events (
    "id" text primary key not null,
    "type" text not null,
    "origin" text not null,
    "payload" jsonb not null default '{}',
    "instance_id" text not null,
    "machine_id" text not null references machines(id) on delete cascade,
    "status" text not null,
    "timestamp" timestamp not null
);

CREATE INDEX machine_events_machine_id_idx ON machine_events(machine_id);
`

const initialDown = `
DROP TABLE machine_events;
DROP TABLE machine_versions;
DROP TABLE machines;
DROP TABLE fleets;
DROP TABLE namespaces;
`
