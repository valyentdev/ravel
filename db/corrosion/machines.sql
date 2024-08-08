CREATE TABLE machines (
    id text primary key not null default "",
    node text not null default "",
    namespace text not null default "",
    fleet_id text not null default "",
    instance_id text not null default "",
    region text not null default "",
    created_at integer not null default 0,
    updated_at integer not null default 0,
    destroyed boolean not null default FALSE
);

CREATE index machines_node on machines(node);
CREATE index machines_instance_id on machines(instance_id);
CREATE index machines_fleet_id on machines(fleet_id);
CREATE index machines_namespace on machines(namespace);
CREATE index machines_region on machines(region);