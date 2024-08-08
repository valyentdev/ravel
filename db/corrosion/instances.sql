CREATE TABLE instances (
    id text not null default '',
    machine_id text not null default '',
    node text not null default '',
    config jsonb not null default '{}',
    image_ref text not null default '',
    local_ip_v4 text not null default '',
    created_at integer not null default 0,
    updated_at integer not null default 0,
    primary key (id, machine_id)
);

CREATE TABLE instance_statuses (
    id text not null,
    machine_id text not null default '',
    status text not null default '',
    updated_at integer not null default 0,
    primary key (id, machine_id)
);

CREATE index instances_machine_id on instances(machine_id);
CREATE index instances_node on instances(node);
CREATE index instance_statuses_machine_id on instance_statuses(machine_id);
