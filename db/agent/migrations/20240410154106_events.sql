-- migrate:up


CREATE TABLE instance_events (
    "id" text primary key not null,
    "type" text not null,
    "origin" text not null,
    "payload" jsonb not null,
    "instance_id" text not null references instances(id) on delete cascade,
    "status" text not null,
    "should_report" boolean not null,
    "reported" boolean not null,
    "timestamp" timestamp not null
);

-- migrate:down

DROP TABLE machine_events;
