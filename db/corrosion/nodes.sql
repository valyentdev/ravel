CREATE TABLE nodes (
    id text primary key not null default '',
    address text not null default '',
    agent_port integer not null default 0,
    http_proxy_port integer not null default 0,
    region text not null default '',
    heartbeated_at integer not null default 0
);

CREATE index nodes_address on nodes(address);
CREATE index nodes_region on nodes(region);