-- migrate:up

CREATE TABLE instance_events (
    "id" text primary key not null,
    "type" text not null,
    "origin" text not null,
    "payload" jsonb not null,
    "instance_id" text not null,
    "machine_id" text not null REFERENCES machines(id) ON DELETE CASCADE,
    "status" text not null,
    "timestamp" timestamp not null
);

CREATE INDEX instance_events_machine_id_idx ON instance_events(machine_id);

-- migrate:down

DROP INDEX instance_events_instance_id_idx;
DROP INDEX instance_events_machine_id_idx;
DROP TABLE instance_events;