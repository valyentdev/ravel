CREATE table gateways (
    id text not null primary key default '',
    namespace text not null default '',
    fleet_id text not null default '',
    name text not null default '',
    service text not null default '',
    protocol text not null default 'https',
    target_port int not null default 80
)