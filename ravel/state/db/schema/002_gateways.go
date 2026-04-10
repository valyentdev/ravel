package schema

const gatewaysUp = `
CREATE TABLE gateways (
    "id" text primary key,
    "name" text not null,
    "fleet_id" text not null references fleets("id")  on delete cascade,
    "namespace" text not null references namespaces("name") on delete cascade,
    "target_port" integer not null,
    "protocol" text not null default '',
	CONSTRAINT unique_gateway_name_in_namespace UNIQUE (namespace, name)
);
CREATE INDEX gateways_fleet_id_idx ON gateways(fleet_id);
CREATE INDEX gateways_namespace_idx ON gateways(namespace);
`

const gatewaysDown = `
DROP TABLE gateways;
`
