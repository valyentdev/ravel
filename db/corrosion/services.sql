CREATE TABLE instance_services (
    service_id text not null default '',
    instance_id text not null default '',
    protocol text not null default 'http',
    port integer not null default 0,
    local_ipv4_address text not null default '',
    primary key (service_id, instance_id)
)