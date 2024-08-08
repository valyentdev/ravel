-- migrate:up

CREATE table reservations (
    "id" text primary key not null,
    "cpus" integer not null,
    "memory" integer not null,
    "local_ipv4_subnet" text unique not null,
    "created_at" timestamp not null default current_timestamp,
    "status" text not null
);


-- migrate:down

DROP TABLE reservations;