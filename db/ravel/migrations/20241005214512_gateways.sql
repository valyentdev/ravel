-- migrate:up

CREATE TABLE gateways (
    id text primary key,
    name TEXT NOT NULL,
    fleet_id TEXT NOT NULL REFERENCES fleets(id),
    namespace TEXT NOT NULL REFERENCES namespaces(name),
    target_port INT NOT NULL,
    protocol TEXT NOT NULL
);

CREATE UNIQUE INDEX gateways_name_idx ON gateways(namespace, name);


-- migrate:down


DROP TABLE gateways;
