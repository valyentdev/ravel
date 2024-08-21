CREATE TABLE instances (
    id text not null default '',
    node text not null default '',
    machine_id text not null default '',
    machine_version text not null default '',
    status text not null default '',
    created_at integer not null default 0,
    updated_at integer not null default 0,
    primary key (id, machine_id)
);

CREATE index instances_machine_id on instances(machine_id);
CREATE index instances_node on instances(node);
