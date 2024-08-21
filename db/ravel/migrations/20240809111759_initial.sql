-- migrate:up


CREATE TABLE  namespaces (
    name text primary key,
    created_at timestamp not null default timezone('utc', now())
);

CREATE TABLE fleets (
    id text primary key,
    name text not null,
    namespace text not null references namespaces(name),
    created_at timestamp not null,
    status text not null default 'active'
);

CREATE UNIQUE INDEX fleets_name_idx ON  fleets USING btree(namespace, name) WHERE destroyed = FALSE;
CREATE INDEX fleets_namespace_idx ON fleets(namespace);
CREATE INDEX fleets_namespace_name_idx ON fleets(namespace, name);

CREATE TABLE machines (
    id text primary key,
    namespace text not null references namespaces(name),
    fleet_id text not null references fleets(id),
    instance_id text not null default "",
    machine_version text not null default "",
    node text not null default "",
    region text not null default "",
    created_at timestamp not null default timezone('utc', now()),
    updated_at timestamp not null default timezone('utc', now()),
    destroyed boolean not null default FALSE
);


CREATE TABLE machine_versions (
    id text primary key,
    machine_id text not null references machines(id) on delete cascade,
    config jsonb not null default '{}',
);


-- migrate:down


DROP TABLE machines;
DROP TABLE fleet;
DROP TABLE namespaces;
